package image

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/terakoya76/shrinkr/internal/runner"
)

var ErrNoCWebP = errors.New("cwebp not installed; skipping WebP file")

type WebPOptions struct {
	Quality int
}

// EncodeWebP re-encodes a WebP via cwebp. There is no pure-Go WebP encoder in
// the standard library, so if cwebp is missing we return ErrNoCWebP and the
// caller reports it as a skip.
func EncodeWebP(ctx context.Context, r runner.Runner, src, dst string, opts WebPOptions) error {
	if _, err := r.LookPath("cwebp"); err != nil {
		return ErrNoCWebP
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := r.Run(ctx, "cwebp",
		"-q", fmt.Sprintf("%d", opts.Quality),
		"-m", "6",
		"-quiet",
		src, "-o", dst,
	)
	if err != nil {
		return fmt.Errorf("cwebp: %w: %s", err, string(out))
	}
	return nil
}
