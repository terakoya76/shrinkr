package video

import (
	"slices"
	"testing"
)

func TestCodecArgs(t *testing.T) {
	cases := []struct {
		source     string
		wantArgs   []string
		wantOffset int
	}{
		{"hevc", []string{"-c:v", "libx265", "-tag:v", "hvc1"}, 5},
		{"h265", []string{"-c:v", "libx265", "-tag:v", "hvc1"}, 5},
		{"h264", []string{"-c:v", "libx264"}, 0},
		{"vp9", []string{"-c:v", "libx264"}, 0},
		{"", []string{"-c:v", "libx264"}, 0},
	}
	for _, c := range cases {
		gotArgs, gotOffset := codecArgs(c.source)
		if !slices.Equal(gotArgs, c.wantArgs) {
			t.Errorf("codecArgs(%q) args = %v, want %v", c.source, gotArgs, c.wantArgs)
		}
		if gotOffset != c.wantOffset {
			t.Errorf("codecArgs(%q) offset = %d, want %d", c.source, gotOffset, c.wantOffset)
		}
	}
}
