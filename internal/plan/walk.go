package plan

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

type Job struct {
	SrcPath string
	DstPath string
	RelPath string
	Kind    Kind
	SrcSize int64
	SrcMod  int64
}

type WalkOptions struct {
	Src         string
	Dst         string
	IncludeGlob string
	ExcludeGlob string
}

// Walk enumerates Jobs for every media file under Src, mirroring the tree
// into Dst. Non-media files and dirs matching skip patterns are ignored.
func Walk(opts WalkOptions) ([]Job, error) {
	src := filepath.Clean(opts.Src)
	dst := filepath.Clean(opts.Dst)
	var jobs []Job

	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if opts.IncludeGlob != "" {
			ok, err := filepath.Match(opts.IncludeGlob, filepath.Base(path))
			if err != nil {
				return fmt.Errorf("bad include glob: %w", err)
			}
			if !ok {
				return nil
			}
		}
		if opts.ExcludeGlob != "" {
			ok, err := filepath.Match(opts.ExcludeGlob, filepath.Base(path))
			if err != nil {
				return fmt.Errorf("bad exclude glob: %w", err)
			}
			if ok {
				return nil
			}
		}

		// Skip Google Takeout sidecar JSONs and album metadata.
		base := filepath.Base(path)
		if strings.HasSuffix(base, ".json") || base == "metadata.json" {
			return nil
		}

		kind, err := Classify(path)
		if err != nil {
			return nil
		}
		if kind == KindUnknown {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		jobs = append(jobs, Job{
			SrcPath: path,
			DstPath: filepath.Join(dst, dstRel(rel, kind)),
			RelPath: rel,
			Kind:    kind,
			SrcSize: info.Size(),
			SrcMod:  info.ModTime().Unix(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return jobs, nil
}

// dstRel replaces the extension when the output format differs from input.
// HEIC always becomes JPEG and video files always become MP4 in the output
// tree, so downstream consumers can round-trip via Google Photos safely.
func dstRel(rel string, kind Kind) string {
	switch kind {
	case KindHEIC:
		ext := filepath.Ext(rel)
		return strings.TrimSuffix(rel, ext) + ".jpg"
	case KindVideo:
		ext := filepath.Ext(rel)
		if strings.EqualFold(ext, ".mp4") {
			return rel
		}
		return strings.TrimSuffix(rel, ext) + ".mp4"
	}
	return rel
}
