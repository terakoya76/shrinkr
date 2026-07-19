package plan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClassifyByExtension(t *testing.T) {
	cases := []struct {
		name string
		want Kind
	}{
		{"IMG_0001.jpg", KindJPEG},
		{"IMG_0001.JPEG", KindJPEG},
		{"scan.PNG", KindPNG},
		{"live.HEIC", KindHEIC},
		{"animated.webp", KindWebP},
		{"clip.mp4", KindVideo},
		{"clip.MOV", KindVideo},
		{"clip.mkv", KindVideo},
		{"notes.txt", KindUnknown},
	}
	dir := t.TempDir()
	for _, c := range cases {
		path := filepath.Join(dir, c.name)
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		got, err := Classify(path)
		if err != nil {
			if c.want == KindUnknown {
				continue
			}
			t.Errorf("%s: err %v", c.name, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s: kind=%s want %s", c.name, got, c.want)
		}
	}
}

func TestClassifyJPEGByMagicBytes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "no-extension")
	// JPEG SOI + APP0
	data := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	kind, err := Classify(path)
	if err != nil {
		t.Fatal(err)
	}
	if kind != KindJPEG {
		t.Errorf("kind=%s want jpeg", kind)
	}
}
