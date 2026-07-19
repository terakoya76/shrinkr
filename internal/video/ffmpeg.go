package video

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/terakoya76/shrinkr/internal/runner"
)

var ErrNoFFmpeg = errors.New("ffmpeg not installed; skipping video file")

type Options struct {
	MaxHeight int
	CRF       int
}

// Encode re-encodes src into H.264 or H.265, downscaling to at most MaxHeight
// and preserving container-level metadata (creation_time, etc.). The output is
// dst.
//
// The output codec is picked to match the source: HEVC input stays HEVC
// (libx265), everything else becomes H.264 (libx264). That matters because
// H.264 is a less efficient codec than HEVC — transcoding a well-tuned HEVC
// source to H.264 at the same CRF often produces a file larger than the
// original, which is the opposite of what a compression tool should do.
func Encode(ctx context.Context, r runner.Runner, src, dst string, opts Options) error {
	if _, err := r.LookPath("ffmpeg"); err != nil {
		return ErrNoFFmpeg
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	// Errors are non-fatal here: an unavailable or misbehaving ffprobe drops us
	// into the H.264 branch, which is fine — we just miss the HEVC-preserving
	// path for this file.
	info, _ := Probe(ctx, r, src)

	codecOpts, crfOffset := codecArgs(info.VideoCodec)
	scale := fmt.Sprintf("scale=-2:'min(%d,ih)'", opts.MaxHeight)
	args := []string{
		"-y",
		"-loglevel", "error",
		"-i", src,
		"-map_metadata", "0",
		"-movflags", "+use_metadata_tags+faststart",
	}
	args = append(args, codecOpts...)
	args = append(args,
		"-preset", "medium",
		"-crf", fmt.Sprintf("%d", opts.CRF+crfOffset),
		"-pix_fmt", "yuv420p",
		"-vf", scale,
		"-c:a", "aac",
		"-b:a", "128k",
		dst,
	)
	out, err := r.Run(ctx, "ffmpeg", args...)
	if err != nil {
		return fmt.Errorf("ffmpeg: %w: %s", err, string(out))
	}
	return nil
}

// codecArgs returns the ffmpeg -c:v flags for the chosen output codec, and a
// CRF offset that maps the caller's H.264-shaped CRF value onto that codec.
//
// HEVC → libx265 tagged as hvc1 (broad player compatibility), CRF + 5.
// The offset reflects that libx265 CRF N is roughly equivalent in visual
// quality to libx264 CRF (N - 5); without it, libx265 with a "balanced"
// CRF 26 targets near-source visual quality and produces almost no size win
// (and can grow the file) when the source is already high-bitrate HEVC.
//
// Everything else → libx264, no offset.
func codecArgs(sourceCodec string) (args []string, crfOffset int) {
	switch sourceCodec {
	case "hevc", "h265":
		return []string{"-c:v", "libx265", "-tag:v", "hvc1"}, 5
	default:
		return []string{"-c:v", "libx264"}, 0
	}
}
