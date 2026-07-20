package image

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"golang.org/x/image/draw"
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
	var newW, newH int
	if w >= h {
		newW = maxEdge
		newH = maxEdge * h / w
	} else {
		newH = maxEdge
		newW = maxEdge * w / h
	}
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	// CatmullRom (bicubic) is visually indistinguishable from Lanczos for the
	// downscale ratios we hit (photo max-edge caps), and lets us drop the
	// disintegration/imaging dependency which is unmaintained and carries an
	// open (unreachable-for-us) TIFF-handling CVE.
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, b, draw.Src, nil)
	return dst
}
