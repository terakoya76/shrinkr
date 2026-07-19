package image

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/disintegration/imaging"
)

// LoadAndFit reads an image and returns a version scaled so that the longer
// edge is at most maxEdge. If the image is already smaller, it is returned
// unchanged.
func LoadAndFit(path string, maxEdge int) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return fit(img, maxEdge), nil
}

func fit(img image.Image, maxEdge int) image.Image {
	if maxEdge <= 0 {
		return img
	}
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= maxEdge && h <= maxEdge {
		return img
	}
	if w >= h {
		return imaging.Resize(img, maxEdge, 0, imaging.Lanczos)
	}
	return imaging.Resize(img, 0, maxEdge, imaging.Lanczos)
}
