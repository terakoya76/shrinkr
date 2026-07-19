package plan

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestWalkMirrorsTreeAndRewritesHEICandVideo(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	mustWrite := func(rel, ext string) {
		p := filepath.Join(src, rel+ext)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mustWrite("a/photo1", ".jpg")
	mustWrite("a/photo2", ".heic")
	mustWrite("b/clip1", ".mov")
	mustWrite("b/clip2", ".mp4")
	mustWrite("b/metadata", ".json") // must be skipped

	jobs, err := Walk(WalkOptions{Src: src, Dst: dst})
	if err != nil {
		t.Fatal(err)
	}
	got := make([]string, len(jobs))
	for i, j := range jobs {
		rel, _ := filepath.Rel(dst, j.DstPath)
		got[i] = rel
	}
	sort.Strings(got)
	want := []string{
		filepath.Join("a", "photo1.jpg"),
		filepath.Join("a", "photo2.jpg"),
		filepath.Join("b", "clip1.mp4"),
		filepath.Join("b", "clip2.mp4"),
	}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("dst[%d]=%q want %q", i, got[i], want[i])
		}
	}
}
