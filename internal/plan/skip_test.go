package plan

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestShouldSkip(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "out.jpg")

	job := Job{DstPath: dst, SrcMod: time.Now().Unix()}

	if ShouldSkip(job, false) {
		t.Errorf("skip when dst does not exist should be false")
	}

	if err := os.WriteFile(dst, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	future := time.Now().Add(1 * time.Hour)
	if err := os.Chtimes(dst, future, future); err != nil {
		t.Fatal(err)
	}
	if !ShouldSkip(job, false) {
		t.Errorf("skip when dst is newer should be true")
	}

	if ShouldSkip(job, true) {
		t.Errorf("overwrite=true should bypass skip")
	}
}
