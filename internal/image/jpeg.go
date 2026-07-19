package image

import (
	"bufio"
	"context"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"os"
	"path/filepath"

	"github.com/terakoya76/shrinkr/internal/runner"
)

type JPEGOptions struct {
	MaxEdge int
	Quality int
}

// EncodeJPEG produces a compressed JPEG at dst from src. When cjpeg (mozjpeg
// or libjpeg-turbo) is available it is preferred because it consistently
// yields 15-30% smaller output at the same visual quality. Otherwise the
// pure-Go image/jpeg encoder is used.
func EncodeJPEG(ctx context.Context, r runner.Runner, src, dst string, opts JPEGOptions) error {
	img, err := LoadAndFit(src, opts.MaxEdge)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if _, err := r.LookPath("cjpeg"); err == nil {
		return encodeViaMozJPEG(ctx, r, img, dst, opts.Quality)
	}
	return encodeViaStdlibJPEG(img, dst, opts.Quality)
}

func encodeViaMozJPEG(ctx context.Context, r runner.Runner, img image.Image, dst string, quality int) error {
	// Hand cjpeg a PPM (P6) intermediate rather than a PNG: PNG input was only
	// added to libjpeg-turbo's cjpeg in v3.0, and Ubuntu 24.04 still ships 2.1.5.
	// PPM has been accepted by every cjpeg for decades.
	tmpPath, err := writeTempPPM(img)
	if err != nil {
		return err
	}
	defer os.Remove(tmpPath)

	out, err := r.Run(ctx, "cjpeg",
		"-quality", fmt.Sprintf("%d", quality),
		"-optimize",
		"-outfile", dst,
		tmpPath,
	)
	if err != nil {
		return fmt.Errorf("cjpeg: %w: %s", err, string(out))
	}
	return nil
}

func writeTempPPM(img image.Image) (string, error) {
	tmp, err := os.CreateTemp("", "shrinkr-*.ppm")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()

	b := img.Bounds()
	rgba := image.NewRGBA(b)
	draw.Draw(rgba, b, img, b.Min, draw.Src)

	w := bufio.NewWriter(tmp)
	if _, err := fmt.Fprintf(w, "P6\n%d %d\n255\n", b.Dx(), b.Dy()); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return "", err
	}
	for i := 0; i < len(rgba.Pix); i += 4 {
		if _, err := w.Write(rgba.Pix[i : i+3]); err != nil {
			tmp.Close()
			os.Remove(tmpPath)
			return "", err
		}
	}
	if err := w.Flush(); err != nil {
		tmp.Close()
		os.Remove(tmpPath)
		return "", err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return "", err
	}
	return tmpPath, nil
}

func encodeViaStdlibJPEG(img image.Image, dst string, quality int) error {
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, img, &jpeg.Options{Quality: quality})
}
