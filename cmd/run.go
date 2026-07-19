package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/schollz/progressbar/v3"

	"github.com/terakoya76/shrinkr/internal/config"
	"github.com/terakoya76/shrinkr/internal/image"
	"github.com/terakoya76/shrinkr/internal/plan"
	"github.com/terakoya76/shrinkr/internal/report"
	"github.com/terakoya76/shrinkr/internal/runner"
	"github.com/terakoya76/shrinkr/internal/video"
	"github.com/terakoya76/shrinkr/internal/worker"
)

type RunCmd struct {
	Src            string  `arg:"" help:"Source directory (Takeout album)."`
	Dst            string  `arg:"" help:"Destination directory for compressed output."`
	Preset         string  `default:"balanced" help:"Preset name (balanced/aggressive/conservative)."`
	Config         string  `help:"Path to YAML config file overlaid on the preset."`
	Workers        int     `default:"0" help:"Concurrent jobs (0 = NumCPU)."`
	DryRun         bool    `help:"Plan and report but do not write output."`
	MinSavings     float64 `default:"0" help:"Skip write if savings ratio is below this (0-1)."`
	ImageMaxEdge   int     `default:"0" help:"Long-edge cap for images (0 = preset)."`
	VideoMaxHeight int     `default:"0" help:"Height cap for videos (0 = preset)."`
	JPEGQuality    int     `default:"0" help:"JPEG quality (0 = preset)."`
	WebPQuality    int     `name:"webp-quality" default:"0" help:"WebP quality (0 = preset)."`
	VideoCRF       int     `default:"0" help:"H.264 CRF (0 = preset)."`
	ReportPath     string  `name:"report" help:"Path to write JSON summary."`
	Overwrite      bool    `help:"Overwrite existing outputs even if they look current."`
	IncludeGlob    string  `help:"Only process filenames matching this glob."`
	ExcludeGlob    string  `help:"Skip filenames matching this glob."`
}

func (c *RunCmd) Run(ctx Context) error {
	presets, err := config.LoadPresets()
	if err != nil {
		return err
	}
	preset, ok := presets[c.Preset]
	if !ok {
		return fmt.Errorf("unknown preset %q", c.Preset)
	}
	merged, err := config.Merge(preset, c.Config, config.Overrides{
		ImageMaxEdge:   c.ImageMaxEdge,
		JPEGQuality:    c.JPEGQuality,
		WebPQuality:    c.WebPQuality,
		VideoMaxHeight: c.VideoMaxHeight,
		VideoCRF:       c.VideoCRF,
		MinSavings:     c.MinSavings,
	})
	if err != nil {
		return err
	}

	jobs, err := plan.Walk(plan.WalkOptions{
		Src:         c.Src,
		Dst:         c.Dst,
		IncludeGlob: c.IncludeGlob,
		ExcludeGlob: c.ExcludeGlob,
	})
	if err != nil {
		return fmt.Errorf("walk: %w", err)
	}
	if len(jobs) == 0 {
		fmt.Println("no media files found under", c.Src)
		return nil
	}

	fmt.Printf("shrinkr: %d jobs, preset=%s, workers=%d, dry-run=%v\n",
		len(jobs), c.Preset, effectiveWorkers(c.Workers), c.DryRun)

	if c.DryRun {
		for _, j := range jobs {
			fmt.Printf("  %s  %s -> %s\n", j.Kind, j.SrcPath, j.DstPath)
		}
		return nil
	}

	r := runner.Exec{}
	rec := report.New()
	bar := progressbar.NewOptions(len(jobs),
		progressbar.OptionSetDescription("compressing"),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
	)

	handler := func(hctx context.Context, job plan.Job) worker.Result {
		return processJob(hctx, r, job, merged)
	}
	worker.Run(ctx.Ctx, jobs, effectiveWorkers(c.Workers), handler, func(res worker.Result) {
		rec.Add(res)
		_ = bar.Add(1)
	})
	_ = bar.Finish()
	fmt.Println()

	if c.ReportPath != "" {
		if err := rec.WriteJSON(c.ReportPath); err != nil {
			return fmt.Errorf("write report: %w", err)
		}
	}
	rec.WriteHuman(os.Stdout)

	s := rec.Summary()
	if s.TotalJobs > 0 && s.Failed == s.TotalJobs {
		return fmt.Errorf("all %d jobs failed", s.Failed)
	}
	return nil
}

func effectiveWorkers(n int) int {
	if n > 0 {
		return n
	}
	return runtime.NumCPU()
}

func processJob(ctx context.Context, r runner.Runner, job plan.Job, p config.Preset) worker.Result {
	if plan.ShouldSkip(job, false) {
		return worker.Result{Job: job, Action: worker.ActionSkipped, Reason: "output current"}
	}
	var err error
	switch job.Kind {
	case plan.KindJPEG:
		err = image.EncodeJPEG(ctx, r, job.SrcPath, job.DstPath, image.JPEGOptions{
			MaxEdge: p.ImageMaxEdge, Quality: p.JPEGQuality,
		})
	case plan.KindPNG:
		err = image.OptimizePNG(ctx, r, job.SrcPath, job.DstPath)
	case plan.KindHEIC:
		err = image.EncodeHEICtoJPEG(ctx, r, job.SrcPath, job.DstPath, image.JPEGOptions{
			MaxEdge: p.ImageMaxEdge, Quality: p.JPEGQuality,
		})
	case plan.KindWebP:
		err = image.EncodeWebP(ctx, r, job.SrcPath, job.DstPath, image.WebPOptions{
			Quality: p.WebPQuality,
		})
	case plan.KindVideo:
		err = video.Encode(ctx, r, job.SrcPath, job.DstPath, video.Options{
			MaxHeight: p.VideoMaxHeight, CRF: p.VideoCRF,
		})
	}
	if err != nil {
		if errors.Is(err, image.ErrNoCWebP) ||
			errors.Is(err, image.ErrNoHEIFConvert) ||
			errors.Is(err, video.ErrNoFFmpeg) {
			return worker.Result{Job: job, Action: worker.ActionSkipped, Reason: err.Error()}
		}
		_ = os.Remove(job.DstPath)
		return worker.Result{Job: job, Action: worker.ActionFailed, Err: err}
	}
	if err := image.CopyMetadata(ctx, r, job.SrcPath, job.DstPath); err != nil {
		return worker.Result{Job: job, Action: worker.ActionFailed, Err: fmt.Errorf("copy metadata: %w", err)}
	}
	dstInfo, statErr := os.Stat(job.DstPath)
	if statErr != nil {
		return worker.Result{Job: job, Action: worker.ActionFailed, Err: statErr}
	}
	dstSize := dstInfo.Size()

	if job.SrcSize > 0 && p.MinSavings > 0 {
		saved := float64(job.SrcSize-dstSize) / float64(job.SrcSize)
		if saved < p.MinSavings {
			_ = os.Remove(job.DstPath)
			if err := copyOriginal(job.SrcPath, job.DstPath); err != nil {
				return worker.Result{Job: job, Action: worker.ActionFailed, Err: err}
			}
			return worker.Result{
				Job: job, Action: worker.ActionCopied, DstSize: job.SrcSize,
				Reason: fmt.Sprintf("savings %.1f%% below threshold; kept original", saved*100),
			}
		}
	}
	return worker.Result{Job: job, Action: worker.ActionCompressed, DstSize: dstSize}
}

func copyOriginal(src, dst string) error {
	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()
	df, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer df.Close()
	if _, err := io.Copy(df, sf); err != nil {
		return err
	}
	return nil
}
