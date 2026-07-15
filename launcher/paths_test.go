package main

import (
	"path/filepath"
	"testing"
)

func TestRuntimePathsAreRelativeToExecutableDirectory(t *testing.T) {
	appDir := filepath.Join(`C:\Program Files`, "MINIMTS")

	corePath, opencvDir := runtimePaths(appDir)

	if want := filepath.Join(appDir, "minimtspro-core.exe"); corePath != want {
		t.Errorf("core path = %q, want %q", corePath, want)
	}
	if want := filepath.Join(appDir, "opencv"); opencvDir != want {
		t.Errorf("OpenCV directory = %q, want %q", opencvDir, want)
	}
}
