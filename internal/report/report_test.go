package report

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/terakoya76/shrinkr/internal/plan"
	"github.com/terakoya76/shrinkr/internal/worker"
)

func TestRecorderSummary(t *testing.T) {
	r := New()
	r.Add(worker.Result{
		Job:     plan.Job{RelPath: "a.jpg", Kind: plan.KindJPEG, SrcSize: 1_000_000},
		Action:  worker.ActionCompressed,
		DstSize: 300_000,
	})
	r.Add(worker.Result{
		Job:     plan.Job{RelPath: "b.mp4", Kind: plan.KindVideo, SrcSize: 5_000_000},
		Action:  worker.ActionSkipped,
		DstSize: 0,
		Reason:  "output current",
	})
	s := r.Summary()
	if s.TotalJobs != 2 {
		t.Errorf("TotalJobs=%d want 2", s.TotalJobs)
	}
	if s.Compressed != 1 || s.Skipped != 1 {
		t.Errorf("action counts wrong: %+v", s)
	}
	if s.SrcBytes != 1_000_000 || s.DstBytes != 300_000 {
		t.Errorf("bytes wrong: %+v", s)
	}
	if got := s.SavedRatio; got < 0.69 || got > 0.71 {
		t.Errorf("SavedRatio=%f want ~0.70", got)
	}

	var buf bytes.Buffer
	r.WriteHuman(&buf)
	if !strings.Contains(buf.String(), "compressed: 1") {
		t.Errorf("human output missing compressed line: %s", buf.String())
	}
}

func TestRecorderWriteJSON(t *testing.T) {
	r := New()
	r.Add(worker.Result{
		Job:     plan.Job{RelPath: "a.jpg", Kind: plan.KindJPEG, SrcSize: 100},
		Action:  worker.ActionCompressed,
		DstSize: 40,
	})
	path := filepath.Join(t.TempDir(), "report.json")
	if err := r.WriteJSON(path); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	var s Summary
	raw := readAll(t, path)
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(s.Entries) != 1 || s.Entries[0].Rel != "a.jpg" {
		t.Errorf("entries wrong: %+v", s.Entries)
	}
}

func readAll(t *testing.T, path string) []byte {
	t.Helper()
	b := new(bytes.Buffer)
	f := openFile(t, path)
	defer f.Close()
	_, err := b.ReadFrom(f)
	if err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
