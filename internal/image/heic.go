package image

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/terakoya76/shrinkr/internal/runner"
)

var ErrNoHEIFConvert = errors.New("heif-convert not installed; skipping HEIC file")

// EncodeHEICtoJPEG converts HEIC to JPEG via heif-convert then re-encodes with
// EncodeJPEG. Google Photos handles JPEG identically to HEIC for timeline
// placement as long as EXIF DateTimeOriginal is preserved.
func EncodeHEICtoJPEG(ctx context.Context, r runner.Runner, src, dst string, opts JPEGOptions) error {
	if _, err := r.LookPath("heif-convert"); err != nil {
		return ErrNoHEIFConvert
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp("", "shrinkr-*.jpg")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpPath)

	out, err := r.Run(ctx, "heif-convert", "-q", "95", src, tmpPath)
	if err != nil {
		return fmt.Errorf("heif-convert: %w: %s", err, string(out))
	}
	return EncodeJPEG(ctx, r, tmpPath, dst, opts)
}
