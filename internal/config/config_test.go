package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPresetsHasCoreThree(t *testing.T) {
	presets, err := LoadPresets()
	if err != nil {
		t.Fatalf("LoadPresets: %v", err)
	}
	for _, name := range []string{"aggressive", "balanced", "conservative"} {
		if _, ok := presets[name]; !ok {
			t.Errorf("missing preset %q", name)
		}
	}
}

func TestMergeFlagsOverrideConfigOverrideBase(t *testing.T) {
	base := Preset{ImageMaxEdge: 4096, JPEGQuality: 82, VideoCRF: 26, MinSavings: 0.10}

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "overlay.yaml")
	if err := os.WriteFile(cfgPath, []byte("image_max_edge: 3000\njpeg_quality: 88\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Merge(base, cfgPath, Overrides{JPEGQuality: 70})
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if got.ImageMaxEdge != 3000 {
		t.Errorf("ImageMaxEdge=%d, want 3000 (from config)", got.ImageMaxEdge)
	}
	if got.JPEGQuality != 70 {
		t.Errorf("JPEGQuality=%d, want 70 (flag wins)", got.JPEGQuality)
	}
	if got.VideoCRF != 26 {
		t.Errorf("VideoCRF=%d, want 26 (base retained)", got.VideoCRF)
	}
	if got.MinSavings != 0.10 {
		t.Errorf("MinSavings=%f, want 0.10 (base retained)", got.MinSavings)
	}
}

func TestMergeNoConfigNoFlagsReturnsBase(t *testing.T) {
	base := Preset{ImageMaxEdge: 4096, JPEGQuality: 82}
	got, err := Merge(base, "", Overrides{})
	if err != nil {
		t.Fatalf("Merge: %v", err)
	}
	if got != base {
		t.Errorf("Merge with no overrides changed base: %+v vs %+v", got, base)
	}
}
