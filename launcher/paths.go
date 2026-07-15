package main

import (
	"errors"
	"path/filepath"
	"strings"
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

func childEnvironment(environment []string, opencvDir string) []string {
	result := make([]string, 0, len(environment)+1)
	foundPath := false
	for _, entry := range environment {
		name, value, hasValue := strings.Cut(entry, "=")
		if hasValue && strings.EqualFold(name, "PATH") {
			if foundPath {
				continue
			}
			if value == "" {
				result = append(result, "PATH="+opencvDir)
			} else {
				result = append(result, "PATH="+opencvDir+string(filepath.ListSeparator)+value)
			}
			foundPath = true
			continue
		}
		result = append(result, entry)
	}
	if !foundPath {
		result = append(result, "PATH="+opencvDir)
	}
	return result
}
