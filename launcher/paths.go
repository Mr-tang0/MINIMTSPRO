package main

import (
	"errors"
	"path/filepath"
)

func executableDirectory(executablePath string) (string, error) {
	if executablePath == "" {
		return "", errors.New("executable path is empty")
	}
	return filepath.Dir(executablePath), nil
}

func runtimePaths(appDir string) (string, string) {
	return filepath.Join(appDir, "minimtspro-core.exe"), filepath.Join(appDir, "opencv")
}
