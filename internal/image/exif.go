package image

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/terakoya76/shrinkr/internal/runner"
)

// CopyMetadata copies EXIF/XMP tags from src to dst using exiftool when
// available. If exiftool is not installed the call is a no-op and returns nil.
// dst's mtime is also aligned to src's mtime regardless.
func CopyMetadata(ctx context.Context, r runner.Runner, src, dst string) error {
	if _, err := r.LookPath("exiftool"); err == nil {
		out, err := r.Run(ctx, "exiftool",
			"-TagsFromFile", src,
			"-overwrite_original",
			"-preserve",
			"-q", "-q",
			dst,
		)
		if err != nil {
			return fmt.Errorf("exiftool: %w: %s", err, string(out))
		}
	}
	return alignMtime(src, dst)
}

func alignMtime(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chtimes(dst, time.Now(), info.ModTime())
}
