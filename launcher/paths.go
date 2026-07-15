package main

import "path/filepath"

func runtimePaths(appDir string) (string, string) {
	return filepath.Join(appDir, "minimtspro-core.exe"), filepath.Join(appDir, "opencv")
}
