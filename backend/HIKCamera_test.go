package backend

import (
	"testing"

	"changeme/backend/extensometer"
)

func TestROIPairTrackingAxis(t *testing.T) {
	left := extensometer.Rect2f{X: 10, Y: 40, Width: 20, Height: 20}
	right := extensometer.Rect2f{X: 80, Y: 45, Width: 20, Height: 20}
	if got := roiPairTrackingAxis(left, right); got != extensometer.TrackingAxisVertical {
		t.Fatalf("left-right axis = %v, want vertical", got)
	}

	top := extensometer.Rect2f{X: 40, Y: 10, Width: 20, Height: 20}
	bottom := extensometer.Rect2f{X: 45, Y: 80, Width: 20, Height: 20}
	if got := roiPairTrackingAxis(top, bottom); got != extensometer.TrackingAxisHorizontal {
		t.Fatalf("top-bottom axis = %v, want horizontal", got)
	}
}
