package cmd

import (
	"fmt"

	"github.com/terakoya76/shrinkr/internal/runner"
)

type DoctorCmd struct{}

type depSpec struct {
	name     string
	required bool
	role     string
}

var deps = []depSpec{
	{"ffmpeg", true, "video re-encode"},
	{"ffprobe", false, "video probe (falls back to defaults if missing)"},
	{"exiftool", false, "EXIF copy (without it, images lose DateTimeOriginal/GPS)"},
	{"cjpeg", false, "mozjpeg — smaller JPEGs (falls back to pure-Go)"},
	{"cwebp", false, "WebP re-encode (WebP files are skipped without it)"},
	{"heif-convert", false, "HEIC decode (HEIC files are skipped without it)"},
	{"oxipng", false, "PNG optimizer (PNGs are copied without it)"},
}

func (c *DoctorCmd) Run(ctx Context) error {
	r := runner.Exec{}
	fmt.Println("shrinkr external dependency check")
	fmt.Println()
	anyMissing := false
	for _, d := range deps {
		path, err := r.LookPath(d.name)
		status := "OK  "
		if err != nil {
			if d.required {
				status = "MISS"
				anyMissing = true
			} else {
				status = "opt "
			}
		}
		if path == "" {
			path = "(not found)"
		}
		fmt.Printf("  [%s] %-12s  %-50s  %s\n", status, d.name, d.role, path)
	}
	if anyMissing {
		fmt.Println("\nOne or more required dependencies are missing. Install them and re-run.")
	}
	return nil
}
