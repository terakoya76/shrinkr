package plan

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strings"
)

type Kind int

const (
	KindUnknown Kind = iota
	KindJPEG
	KindPNG
	KindHEIC
	KindWebP
	KindVideo
)

func (k Kind) String() string {
	switch k {
	case KindJPEG:
		return "jpeg"
	case KindPNG:
		return "png"
	case KindHEIC:
		return "heic"
	case KindWebP:
		return "webp"
	case KindVideo:
		return "video"
	default:
		return "unknown"
	}
}

// Classify determines media kind by inspecting the file's magic bytes.
// Content is authoritative — misnamed files (.HEIC that is actually JPEG,
// .MP4 that is actually MOV, ...) would otherwise blow up downstream
// handlers that trust the extension (heif-convert refuses JPEG-in-HEIC,
// exiftool refuses format mismatches, etc.). Extension is used only as
// a fallback for formats sniff() does not recognize.
func Classify(path string) (Kind, error) {
	if k, err := sniff(path); err != nil {
		return KindUnknown, err
	} else if k != KindUnknown {
		return k, nil
	}
	switch ext := strings.ToLower(extOf(path)); ext {
	case ".jpg", ".jpeg":
		return KindJPEG, nil
	case ".png":
		return KindPNG, nil
	case ".heic", ".heif":
		return KindHEIC, nil
	case ".webp":
		return KindWebP, nil
	case ".mp4", ".mov", ".mkv", ".avi", ".3gp", ".m4v", ".webm":
		return KindVideo, nil
	}
	return KindUnknown, nil
}

func extOf(path string) string {
	i := strings.LastIndex(path, ".")
	if i < 0 {
		return ""
	}
	return path[i:]
}

func sniff(path string) (Kind, error) {
	f, err := os.Open(path)
	if err != nil {
		return KindUnknown, err
	}
	defer f.Close()
	buf := make([]byte, 512)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return KindUnknown, err
	}
	head := buf[:n]

	// HEIC/HEIF ftyp box: bytes 4..12 look like "ftypheic" / "ftypmif1" / "ftyphevc"
	if len(head) >= 12 && bytes.Equal(head[4:8], []byte("ftyp")) {
		brand := string(head[8:12])
		switch brand {
		case "heic", "heix", "mif1", "msf1", "hevc":
			return KindHEIC, nil
		case "isom", "iso2", "avc1", "mp41", "mp42", "M4V ":
			return KindVideo, nil
		case "qt  ":
			return KindVideo, nil
		}
	}

	switch http.DetectContentType(head) {
	case "image/jpeg":
		return KindJPEG, nil
	case "image/png":
		return KindPNG, nil
	case "image/webp":
		return KindWebP, nil
	case "video/mp4", "video/webm", "video/avi", "video/x-matroska":
		return KindVideo, nil
	}
	return KindUnknown, nil
}
