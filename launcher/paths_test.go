package main

import (
	"path/filepath"
	"testing"
)

func TestChildEnvironmentPrependsOpenCVDirectoryToCaseInsensitivePath(t *testing.T) {
	opencvDir := filepath.Join(`C:\Program Files`, "MINIMTS", "opencv")
	environment := []string{
		"USERPROFILE=C:\\Users\\operator",
		"Path=C:\\Windows\\System32",
		"MINIMTS_MODE=production",
	}

	got := childEnvironment(environment, opencvDir)
	want := []string{
		"USERPROFILE=C:\\Users\\operator",
		"PATH=" + opencvDir + string(filepath.ListSeparator) + `C:\Windows\System32`,
		"MINIMTS_MODE=production",
	}
	if len(got) != len(want) {
		t.Fatalf("environment length = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("environment[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestExecutableDirectoryRejectsEmptyPath(t *testing.T) {
	_, err := executableDirectory("")
	if err == nil {
		t.Fatal("executableDirectory(\"\") error = nil, want non-nil")
	}
}

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
