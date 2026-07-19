package config

import (
	_ "embed"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

//go:embed presets.yaml
var presetsYAML []byte

type Preset struct {
	ImageMaxEdge   int     `yaml:"image_max_edge"`
	JPEGQuality    int     `yaml:"jpeg_quality"`
	WebPQuality    int     `yaml:"webp_quality"`
	VideoMaxHeight int     `yaml:"video_max_height"`
	VideoCRF       int     `yaml:"video_crf"`
	MinSavings     float64 `yaml:"min_savings"`
}

type presetsFile struct {
	Presets map[string]Preset `yaml:"presets"`
}

type Config struct {
	Preset         string
	ImageMaxEdge   int
	JPEGQuality    int
	WebPQuality    int
	VideoMaxHeight int
	VideoCRF       int
	MinSavings     float64
	Workers        int
	Overwrite      bool
	DryRun         bool
	IncludeGlob    string
	ExcludeGlob    string
	ReportPath     string
}

func LoadPresets() (map[string]Preset, error) {
	var pf presetsFile
	if err := yaml.Unmarshal(presetsYAML, &pf); err != nil {
		return nil, fmt.Errorf("parse embedded presets: %w", err)
	}
	if len(pf.Presets) == 0 {
		return nil, fmt.Errorf("no presets defined in embedded YAML")
	}
	return pf.Presets, nil
}

// Merge applies preset defaults, then overlays optional YAML config file,
// then overlays CLI flags. Flag values that equal their zero value are treated
// as "not set" and don't override — callers must pass explicit non-zero flags.
type Overrides struct {
	ImageMaxEdge   int
	JPEGQuality    int
	WebPQuality    int
	VideoMaxHeight int
	VideoCRF       int
	MinSavings     float64
}

func Merge(preset Preset, configPath string, flags Overrides) (Preset, error) {
	merged := preset

	if configPath != "" {
		raw, err := os.ReadFile(configPath)
		if err != nil {
			return Preset{}, fmt.Errorf("read config %s: %w", configPath, err)
		}
		var overlay Preset
		if err := yaml.Unmarshal(raw, &overlay); err != nil {
			return Preset{}, fmt.Errorf("parse config %s: %w", configPath, err)
		}
		merged = applyOverlay(merged, overlay)
	}

	merged = applyOverlay(merged, Preset{
		ImageMaxEdge:   flags.ImageMaxEdge,
		JPEGQuality:    flags.JPEGQuality,
		WebPQuality:    flags.WebPQuality,
		VideoMaxHeight: flags.VideoMaxHeight,
		VideoCRF:       flags.VideoCRF,
		MinSavings:     flags.MinSavings,
	})
	return merged, nil
}

func applyOverlay(base, overlay Preset) Preset {
	if overlay.ImageMaxEdge > 0 {
		base.ImageMaxEdge = overlay.ImageMaxEdge
	}
	if overlay.JPEGQuality > 0 {
		base.JPEGQuality = overlay.JPEGQuality
	}
	if overlay.WebPQuality > 0 {
		base.WebPQuality = overlay.WebPQuality
	}
	if overlay.VideoMaxHeight > 0 {
		base.VideoMaxHeight = overlay.VideoMaxHeight
	}
	if overlay.VideoCRF > 0 {
		base.VideoCRF = overlay.VideoCRF
	}
	if overlay.MinSavings > 0 {
		base.MinSavings = overlay.MinSavings
	}
	return base
}
