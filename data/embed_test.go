package data

import (
	"encoding/json"
	"testing"
)

// @lat: [[testing#Package-level unit coverage]]
func TestNewsFilterPresetsValid(t *testing.T) {
	if len(NewsFilterPresets) == 0 {
		t.Fatal("NewsFilterPresets is empty")
	}

	var parsed struct {
		Presets []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"presets"`
	}

	if err := json.Unmarshal(NewsFilterPresets, &parsed); err != nil {
		t.Fatalf("failed to parse NewsFilterPresets: %v", err)
	}

	if len(parsed.Presets) == 0 {
		t.Error("expected at least one preset")
	}

	for _, p := range parsed.Presets {
		if p.ID == "" {
			t.Error("preset has empty ID")
		}
		if p.Name == "" {
			t.Errorf("preset %q has empty name", p.ID)
		}
	}
}
