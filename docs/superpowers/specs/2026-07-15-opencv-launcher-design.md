# OpenCV Runtime Launcher Design

## Goal

Package OpenCV DLLs under an `opencv` subdirectory without modifying the
system `PATH`, while preserving `minimtspro.exe` as the user-facing executable.

## Layout

The installed application directory will contain:

```text
minimtspro.exe          Go launcher
minimtspro-core.exe     Wails application
opencv/                 OpenCV runtime DLLs
```

Shortcuts target `minimtspro.exe`.

## Launcher

Create a separate Windows Go module in `launcher/`. Its executable:

1. Resolves its own directory.
2. Adds `<application directory>\\opencv` to the process DLL search path with
   `SetDefaultDllDirectories` and `AddDllDirectory`.
3. Starts `minimtspro-core.exe`, forwarding the original command-line
   arguments and inheriting standard handles.
4. Waits for the core process and exits with its exit code.
5. Displays a native error dialog when the core executable or the OpenCV
   directory is missing, or when process creation fails.

The launcher does not alter the user or machine environment variables.

## Packaging

The Windows NSIS packaging task will build the launcher, rename the Wails
binary to `minimtspro-core.exe` at installation, and install OpenCV DLLs into
`$INSTDIR\\opencv`. The existing Hikvision SDK DLL remains installed beside the
core application because it is linked directly by the core executable.

## Testing

Unit tests cover construction of the core executable and OpenCV directory
paths from an application directory. A Windows build of the launcher verifies
Win32 API bindings and package compilation. NSIS packaging will be verified
after the Wails CGO task propagation issue is corrected.
