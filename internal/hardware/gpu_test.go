package hardware

import (
	"testing"
)

func TestDetectOutputResolutionFallback(t *testing.T) {
	// For a non-existent connector, should fallback to 1920x1080 and return an error
	w, h, err := DetectOutputResolution("NON_EXISTENT_CONNECTOR_XYZ")
	if err == nil {
		t.Errorf("Expected error for non-existent connector, got nil")
	}
	if w != 1920 || h != 1080 {
		t.Errorf("Expected fallback 1920x1080, got %dx%d", w, h)
	}
}

func TestDetectOutputResolutionLiveOrAuto(t *testing.T) {
	outputs := GetConnectedDisplayOutputs()
	for _, out := range outputs {
		w, h, err := DetectOutputResolution(out)
		if err == nil {
			if w <= 0 || h <= 0 {
				t.Errorf("Expected positive resolution for connected output %s, got %dx%d", out, w, h)
			}
		}
	}

	w, h, err := DetectOutputResolution("auto")
	if w <= 0 || h <= 0 {
		t.Errorf("Expected positive resolution, got %dx%d (err: %v)", w, h, err)
	}
}
