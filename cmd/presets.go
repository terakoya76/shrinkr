package cmd

import (
	"fmt"
	"sort"

	"github.com/terakoya76/shrinkr/internal/config"
)

type PresetsCmd struct{}

func (c *PresetsCmd) Run(ctx Context) error {
	presets, err := config.LoadPresets()
	if err != nil {
		return err
	}
	names := make([]string, 0, len(presets))
	for k := range presets {
		names = append(names, k)
	}
	sort.Strings(names)
	fmt.Printf("%-14s %-10s %-10s %-10s %-10s %-10s %-10s\n",
		"preset", "img_edge", "jpeg_q", "webp_q", "vid_h", "vid_crf", "min_save")
	for _, name := range names {
		p := presets[name]
		fmt.Printf("%-14s %-10d %-10d %-10d %-10d %-10d %-10.2f\n",
			name, p.ImageMaxEdge, p.JPEGQuality, p.WebPQuality, p.VideoMaxHeight, p.VideoCRF, p.MinSavings)
	}
	return nil
}
