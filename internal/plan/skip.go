package plan

import "os"

// ShouldSkip reports whether dst already reflects src (idempotent re-runs).
// A destination is considered current if it exists, is non-empty, and its
// mtime is not older than src's mtime.
func ShouldSkip(job Job, overwrite bool) bool {
	if overwrite {
		return false
	}
	info, err := os.Stat(job.DstPath)
	if err != nil {
		return false
	}
	if info.Size() == 0 {
		return false
	}
	return info.ModTime().Unix() >= job.SrcMod
}
