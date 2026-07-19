// Package report aggregates worker Results into a summary and optional JSON
// output. It is deliberately serial — the caller feeds Results one at a time.
package report

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/terakoya76/shrinkr/internal/worker"
)

type Entry struct {
	Rel     string `json:"rel"`
	Kind    string `json:"kind"`
	Action  string `json:"action"`
	SrcSize int64  `json:"src_size"`
	DstSize int64  `json:"dst_size"`
	Reason  string `json:"reason,omitempty"`
	Error   string `json:"error,omitempty"`
}

type Summary struct {
	TotalJobs   int     `json:"total_jobs"`
	Compressed  int     `json:"compressed"`
	Copied      int     `json:"copied"`
	Skipped     int     `json:"skipped"`
	Failed      int     `json:"failed"`
	SrcBytes    int64   `json:"src_bytes"`
	DstBytes    int64   `json:"dst_bytes"`
	SavedBytes  int64   `json:"saved_bytes"`
	SavedRatio  float64 `json:"saved_ratio"`
	Entries     []Entry `json:"entries"`
}

type Recorder struct {
	summary Summary
}

func New() *Recorder {
	return &Recorder{}
}

func (r *Recorder) Add(res worker.Result) {
	r.summary.TotalJobs++
	e := Entry{
		Rel:     res.Job.RelPath,
		Kind:    res.Job.Kind.String(),
		Action:  res.Action.String(),
		SrcSize: res.Job.SrcSize,
		DstSize: res.DstSize,
		Reason:  res.Reason,
	}
	if res.Err != nil {
		e.Error = res.Err.Error()
	}
	r.summary.Entries = append(r.summary.Entries, e)

	switch res.Action {
	case worker.ActionCompressed:
		r.summary.Compressed++
		r.summary.SrcBytes += res.Job.SrcSize
		r.summary.DstBytes += res.DstSize
	case worker.ActionCopied:
		r.summary.Copied++
		r.summary.SrcBytes += res.Job.SrcSize
		r.summary.DstBytes += res.DstSize
	case worker.ActionSkipped:
		r.summary.Skipped++
	case worker.ActionFailed:
		r.summary.Failed++
	}
}

func (r *Recorder) Summary() Summary {
	s := r.summary
	if s.SrcBytes > 0 {
		s.SavedBytes = s.SrcBytes - s.DstBytes
		s.SavedRatio = float64(s.SavedBytes) / float64(s.SrcBytes)
	}
	return s
}

func (r *Recorder) WriteJSON(path string) error {
	s := r.Summary()
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}

func (r *Recorder) WriteHuman(w io.Writer) {
	s := r.Summary()
	fmt.Fprintf(w, "\n=== shrinkr summary ===\n")
	fmt.Fprintf(w, "  jobs      : %d\n", s.TotalJobs)
	fmt.Fprintf(w, "  compressed: %d\n", s.Compressed)
	fmt.Fprintf(w, "  copied    : %d\n", s.Copied)
	fmt.Fprintf(w, "  skipped   : %d\n", s.Skipped)
	fmt.Fprintf(w, "  failed    : %d\n", s.Failed)
	if s.SrcBytes > 0 {
		fmt.Fprintf(w, "  src bytes : %s\n", humanBytes(s.SrcBytes))
		fmt.Fprintf(w, "  dst bytes : %s\n", humanBytes(s.DstBytes))
		fmt.Fprintf(w, "  saved     : %s (%.1f%%)\n", humanBytes(s.SavedBytes), s.SavedRatio*100)
	}
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}
