package image

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/terakoya76/shrinkr/internal/runner"
)

// OptimizePNG runs oxipng in-place when available; otherwise copies src to dst
// unchanged (PNGs from phone cameras are rare and rarely benefit from pure-Go
// re-encoding).
func OptimizePNG(ctx context.Context, r runner.Runner, src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if err := copyFile(src, dst); err != nil {
		return err
	}
	if _, err := r.LookPath("oxipng"); err != nil {
		return nil
	}
	out, err := r.Run(ctx, "oxipng", "-o", "4", "--strip", "safe", dst)
	if err != nil {
		return fmt.Errorf("oxipng: %w: %s", err, string(out))
	}
	return nil
}

func copyFile(src, dst string) error {
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
	_, err = io.Copy(df, sf)
	return err
}
