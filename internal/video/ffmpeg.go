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

// Encode re-encodes src to H.264 with the requested CRF, downscaling to at most
// MaxHeight and preserving container-level metadata (creation_time, etc.). The
// output is written to dst.
func Encode(ctx context.Context, r runner.Runner, src, dst string, opts Options) error {
	if _, err := r.LookPath("ffmpeg"); err != nil {
		return ErrNoFFmpeg
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	scale := fmt.Sprintf("scale=-2:'min(%d,ih)'", opts.MaxHeight)
	args := []string{
		"-y",
		"-loglevel", "error",
		"-i", src,
		"-map_metadata", "0",
		"-movflags", "+use_metadata_tags+faststart",
		"-c:v", "libx264",
		"-preset", "medium",
		"-crf", fmt.Sprintf("%d", opts.CRF),
		"-pix_fmt", "yuv420p",
		"-vf", scale,
		"-c:a", "aac",
		"-b:a", "128k",
		dst,
	}
	out, err := r.Run(ctx, "ffmpeg", args...)
	if err != nil {
		return fmt.Errorf("ffmpeg: %w: %s", err, string(out))
	}
	return nil
}

