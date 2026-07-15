//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"
)

const (
	loadLibrarySearchDefaultDirs = 0x00001000
	mbOK                         = 0x00000000
	mbIconError                  = 0x00000010
)

var (
	kernel32                     = syscall.NewLazyDLL("kernel32.dll")
	setDefaultDLLDirectoriesProc = kernel32.NewProc("SetDefaultDllDirectories")
	addDLLDirectoryProc          = kernel32.NewProc("AddDllDirectory")
	messageBoxWProc              = syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW")
)

func main() {
	if err := run(); err != nil {
		showError(err.Error())
		os.Exit(1)
	}
}

func run() error {
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate launcher executable: %w", err)
	}
	appDir, err := executableDirectory(executablePath)
	if err != nil {
		return err
	}

	corePath, opencvDir := runtimePaths(appDir)
	if err := requireDirectory(opencvDir); err != nil {
		return err
	}
	if err := requireFile(corePath); err != nil {
		return err
	}
	if err := configureDLLSearch(opencvDir); err != nil {
		return err
	}

	command := exec.Command(corePath, os.Args[1:]...)
	command.Dir = appDir
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		return fmt.Errorf("start core executable: %w", err)
	}
	return nil
}

func requireDirectory(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("OpenCV runtime directory %q: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("OpenCV runtime path %q is not a directory", path)
	}
	return nil
}

func requireFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("core executable %q: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("core executable path %q is a directory", path)
	}
	return nil
}

func configureDLLSearch(opencvDir string) error {
	result, _, callErr := setDefaultDLLDirectoriesProc.Call(loadLibrarySearchDefaultDirs)
	if result == 0 {
		return fmt.Errorf("configure DLL search directories: %w", callErr)
	}

	directory, err := syscall.UTF16PtrFromString(filepath.Clean(opencvDir))
	if err != nil {
		return fmt.Errorf("encode OpenCV runtime directory: %w", err)
	}
	result, _, callErr = addDLLDirectoryProc.Call(uintptr(unsafe.Pointer(directory)))
	if result == 0 {
		return fmt.Errorf("add OpenCV runtime directory: %w", callErr)
	}
	return nil
}

func showError(message string) {
	text, err := syscall.UTF16PtrFromString(message)
	if err != nil {
		return
	}
	title, err := syscall.UTF16PtrFromString("MINIMTS Pro Launcher")
	if err != nil {
		return
	}
	messageBoxWProc.Call(0, uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(title)), mbOK|mbIconError)
}
