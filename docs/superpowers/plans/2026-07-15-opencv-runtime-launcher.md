# OpenCV Runtime Launcher Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Package OpenCV below an `opencv` directory while retaining `minimtspro.exe` as the entry point.

**Architecture:** A Windows Go launcher adds its local `opencv` directory to the process DLL search path before starting the Wails binary as `minimtspro-core.exe`. NSIS installs both executables and the OpenCV runtime in that layout.

**Tech Stack:** Go 1.25 standard library, Win32 kernel32 APIs, Wails 3 Taskfiles, NSIS.

---

### Task 1: Create Testable Launcher Paths

**Files:** Create `launcher/go.mod`, `launcher/paths.go`, and `launcher/paths_test.go`.

- [ ] **Step 1: Write the failing test.** Add `TestRuntimePathsAreRelativeToExecutableDirectory`, passing `C:\\Program Files\\MINIMTS`, and assert `runtimePaths` returns `minimtspro-core.exe` and `opencv` joined to that directory.
- [ ] **Step 2: Confirm the test is red.** Run `go -C launcher test .`; expect failure because the module and `runtimePaths` do not exist.
- [ ] **Step 3: Add minimal code.** Define module `minimtspro-launcher`, Go 1.25.0. Implement `runtimePaths(appDir string) (string, string)` with `filepath.Join(appDir, "minimtspro-core.exe")` and `filepath.Join(appDir, "opencv")`.
- [ ] **Step 4: Confirm the test is green.** Run `go -C launcher test .`; expect `ok minimtspro-launcher`.
- [ ] **Step 5: Commit.** Stage only `launcher/go.mod`, `launcher/paths.go`, and `launcher/paths_test.go`; commit message `feat: add launcher path resolution`.

### Task 2: Implement the Windows Launcher

**Files:** Create `launcher/main_windows.go` and `launcher/main_other.go`; modify `launcher/paths.go` and `launcher/paths_test.go`.

- [ ] **Step 1: Write the failing test.** Add `TestExecutableDirectoryRejectsEmptyPath`; assert `executableDirectory("")` returns an error.
- [ ] **Step 2: Confirm the test is red.** Run `go -C launcher test . -run TestExecutableDirectoryRejectsEmptyPath`; expect undefined `executableDirectory`.
- [ ] **Step 3: Add minimal code.** Implement `executableDirectory` using `filepath.Dir` and reject an empty path. In Windows-only `main`, obtain `os.Executable`, require the core executable and OpenCV directory, call `SetDefaultDllDirectories(LOAD_LIBRARY_SEARCH_DEFAULT_DIRS)` and `AddDllDirectory(opencv)`, then run the core executable with `os.Args[1:]`, attached standard handles, and its exit code. Use `MessageBoxW` and exit 1 for failures. Add an empty non-Windows `main` behind `//go:build !windows`.
- [ ] **Step 4: Compile and test.** Run `go -C launcher test .`, then `Push-Location launcher; go build -o minimtspro.exe .; Pop-Location`; expect passing tests and `launcher/minimtspro.exe`.
- [ ] **Step 5: Commit.** Stage `launcher`; commit message `feat: add Windows OpenCV runtime launcher`.

### Task 3: Integrate Wails and NSIS Packaging

**Files:** Modify `build/windows/Taskfile.yml` and `build/windows/nsis/project.nsi`; create `third_party/opencv/bin/.gitkeep`.

- [ ] **Step 1: Add launcher build task.** Add internal `build:launcher` with directory `{{.ROOT_DIR}}/launcher` and command `go build -o {{.ROOT_DIR}}/{{.BIN_DIR}}/{{.APP_NAME}}.exe .`. Add it as a dependency of `create:nsis:installer`.
- [ ] **Step 2: Install both binaries.** Replace `!insertmacro wails.files` with a renamed Wails payload named `minimtspro-core.exe`, plus the launcher from `bin/minimtspro.exe`. Change output directory to `$INSTDIR\\opencv` for `third_party\\opencv\\bin\\*.dll`; then return to `$INSTDIR` for `MvCameraControl.dll`.
- [ ] **Step 3: Stage the dependency.** Create `third_party\\opencv\\bin`; copy OpenCV runtime DLLs from `C:\\opencv\\build\\install\\x64\\mingw\\bin` and `opencv_videoio_ffmpeg_64.dll` from its ffmpeg directory. Run `wails3 package`.
- [ ] **Step 4: Verify installer output.** In a disposable install directory, confirm the launcher, core binary, `opencv/libopencv_core4130.dll`, and `MvCameraControl.dll`. Run the launcher and require a Wails window with no OpenCV DLL load error.
- [ ] **Step 5: Commit.** Stage the Taskfile, NSIS script, and `.gitkeep`; commit message `build: package OpenCV behind launcher`.
