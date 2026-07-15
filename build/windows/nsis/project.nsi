Unicode true

####
## Please note: Template replacements don't work in this file. They are provided with default defines like
## mentioned underneath.
## If the keyword is not defined, "wails_tools.nsh" will populate them.
## If they are defined here, "wails_tools.nsh" will not touch them. This allows you to use this project.nsi manually
## from outside of Wails for debugging and development of the installer.
##
## For development first make a wails nsis build to populate the "wails_tools.nsh":
## > wails build --target windows/amd64 --nsis
## Then you can call makensis on this file with specifying the path to your binary:
## For a AMD64 only installer:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app.exe
## For a ARM64 only installer:
## > makensis -DARG_WAILS_ARM64_BINARY=..\..\bin\app.exe
## For a installer with both architectures:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app-amd64.exe -DARG_WAILS_ARM64_BINARY=..\..\bin\app-arm64.exe
####
## The following information is taken from the wails_tools.nsh file, but they can be overwritten here.
####
## !define INFO_PROJECTNAME    "my-project" # Default "MINIMTSPRO"
## !define INFO_COMPANYNAME    "My Company" # Default "My Company"
## !define INFO_PRODUCTNAME    "My Product Name" # Default "My Product"
## !define INFO_PRODUCTVERSION "1.0.0"     # Default "0.1.0"
## !define INFO_COPYRIGHT      "(c) Now, My Company" # Default "© 2026, My Company"
###
## !define PRODUCT_EXECUTABLE  "Application.exe"      # Default "${INFO_PROJECTNAME}.exe"
## !define UNINST_KEY_NAME     "UninstKeyInRegistry"  # Default "${INFO_COMPANYNAME}${INFO_PRODUCTNAME}"
####
## !define REQUEST_EXECUTION_LEVEL "admin"            # Default "admin"  see also https://nsis.sourceforge.io/Docs/Chapter4.html
## !define WAILS_INSTALL_SCOPE     "user"             # Default "machine" - set to "user" for per-user install ($LOCALAPPDATA) without UAC prompt
####
## Include the wails tools
####
!include "wails_tools.nsh"

# The version information for this two must consist of 4 parts
VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"

VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

# Enable HiDPI support. https://nsis.sourceforge.io/Reference/ManifestDPIAware
ManifestDPIAware true

!include "MUI.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
# !define MUI_WELCOMEFINISHPAGE_BITMAP "resources\leftimage.bmp" #Include this to add a bitmap on the left side of the Welcome Page. Must be a size of 164x314
!define MUI_FINISHPAGE_NOAUTOCLOSE # Wait on the INSTFILES page so the user can take a look into the details of the installation steps
!define MUI_ABORTWARNING # This will warn the user if they exit from the installer.

!insertmacro MUI_PAGE_WELCOME # Welcome to the installer page.
# !insertmacro MUI_PAGE_LICENSE "resources\eula.txt" # Adds a EULA page to the installer
!insertmacro MUI_PAGE_DIRECTORY # In which folder install page.
!insertmacro MUI_PAGE_INSTFILES # Installing page.
!insertmacro MUI_PAGE_FINISH # Finished installation page.

!insertmacro MUI_UNPAGE_INSTFILES # Uninstalling page

!insertmacro MUI_LANGUAGE "English" # Set the Language of the installer

## The following two statements can be used to sign the installer and the uninstaller. The path to the binaries are provided in %1
#!uninstfinalize 'signtool --file "%1"'
#!finalize 'signtool --file "%1"'

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe" # Name of the installer's file.
!if "${WAILS_INSTALL_SCOPE}" == "user"
    InstallDir "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}"
!else
    InstallDir "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
!endif
ShowInstDetails show # This will always show the installation details.

Function .onInit
   !insertmacro wails.checkArchitecture
FunctionEnd

; ============================================
; Install Section
; ============================================
Section
    !insertmacro wails.setShellContext

    !insertmacro wails.webview2runtime

    ; DPInst requires the INF, catalog, and architecture-specific files next
    ; to the installer executable, so stage the x64 driver package.
    InitPluginsDir
    SetOutPath "$PLUGINSDIR\CP210x_VCP_Windows"
    File "..\..\..\docs\CP210x_VCP_Windows\CP210xVCPInstaller_x64.exe"
    File "..\..\..\docs\CP210x_VCP_Windows\dpinst.xml"
    File "..\..\..\docs\CP210x_VCP_Windows\slabvcp.cat"
    File "..\..\..\docs\CP210x_VCP_Windows\slabvcp.inf"
    SetOutPath "$PLUGINSDIR\CP210x_VCP_Windows\x64"
    File "..\..\..\docs\CP210x_VCP_Windows\x64\silabenm.sys"
    File "..\..\..\docs\CP210x_VCP_Windows\x64\silabser.sys"
    File "..\..\..\docs\CP210x_VCP_Windows\x64\WdfCoInstaller01009.dll"
    ExecWait '"$PLUGINSDIR\CP210x_VCP_Windows\CP210xVCPInstaller_x64.exe" /S /SE' $0
    DetailPrint "CP210x driver installer exited with code $0"

    SetOutPath $INSTDIR

    !insertmacro wails.files

    File "..\..\..\update.json"
    File "..\..\..\.env"

    SetOutPath "$INSTDIR\bin"
    File "C:\msys64\ucrt64\bin\*.dll"
    File "C:\opencv\build\bin\*.dll"

    ; Install the MVS Runtime once per installed application lifecycle.
    ReadRegStr $1 HKLM "Software\PIMS\MINIMTSPRO" "MVSRuntimeV4_8_0_3Installed"
    StrCmp $1 "1" mvsRuntimeDone
    SetOutPath "$PLUGINSDIR\MVS"
    File "..\..\..\docs\MVS_SDK_V4_8_0_3_MVFG_V2_8_0_3_VC90_Runtime_STD.exe"
    ExecWait '"$PLUGINSDIR\MVS\MVS_SDK_V4_8_0_3_MVFG_V2_8_0_3_VC90_Runtime_STD.exe"' $1
    StrCmp $1 0 mvsRuntimeMarkInstalled
    StrCmp $1 3010 mvsRuntimeMarkInstalled
    DetailPrint "MVS Runtime installer exited with code $1"
    Goto mvsRuntimeDone

    mvsRuntimeMarkInstalled:
    WriteRegStr HKLM "Software\PIMS\MINIMTSPRO" "MVSRuntimeV4_8_0_3Installed" "1"
    DetailPrint "MVS Runtime installed"

    mvsRuntimeDone:

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols

    !insertmacro wails.writeUninstaller

    ReadRegStr $0 HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "MINIMTSPRO64"
    StrCmp $0 "C:\Program Files (x86)\Common Files\MVS\Runtime\Win64_x64" pathConfigured
    ReadRegStr $0 HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "Path"
    StrCpy $0 "$0;$INSTDIR\bin;C:\Program Files (x86)\Common Files\MVS\Runtime\Win32_i86;C:\Program Files (x86)\Common Files\MVS\Runtime\Win64_x64"
    WriteRegExpandStr HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "Path" $0

    pathConfigured:
    WriteRegExpandStr HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "MINIMTSPRO32" "C:\Program Files (x86)\Common Files\MVS\Runtime\Win32_i86"
    WriteRegExpandStr HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "MINIMTSPRO64" "C:\Program Files (x86)\Common Files\MVS\Runtime\Win64_x64"
    SendMessage ${HWND_BROADCAST} ${WM_WININICHANGE} 0 "STR:Environment" /TIMEOUT=5000
SectionEnd


; ============================================
; Uninstall Section
; ============================================
Section "uninstall"
    !insertmacro wails.setShellContext

    Call un.removeBinFromPath

    RMDir /r "$AppData\${PRODUCT_EXECUTABLE}" # Remove the WebView2 DataPath

    RMDir /r $INSTDIR

    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols

    !insertmacro wails.deleteUninstaller

    DeleteRegValue HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "MINIMTSPRO32"
    DeleteRegValue HKLM "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" "MINIMTSPRO64"
    DeleteRegValue HKLM "Software\PIMS\MINIMTSPRO" "MVSRuntimeV4_8_0_3Installed"
    DeleteRegKey HKLM "Software\PIMS\MINIMTSPRO"
    SendMessage ${HWND_BROADCAST} ${WM_WININICHANGE} 0 "STR:Environment" /TIMEOUT=5000
SectionEnd

Function un.removeBinFromPath
    InitPluginsDir
    FileOpen $0 "$PLUGINSDIR\remove-minimtspro-bin.ps1" w
    FileWrite $0 "param([string]$$Target)$\r$\n"
    FileWrite $0 "$$key = 'HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment'$\r$\n"
    FileWrite $0 "$$current = (Get-ItemProperty -LiteralPath $$key -Name Path).Path$\r$\n"
    FileWrite $0 "$$items = @($$current -split ';' | Where-Object { $$_ -and $$_ -ine $$Target })$\r$\n"
    FileWrite $0 "Set-ItemProperty -LiteralPath $$key -Name Path -Value ($$items -join ';') -Type ExpandString$\r$\n"
    FileClose $0

    nsExec::ExecToLog '"$SYSDIR\WindowsPowerShell\v1.0\powershell.exe" -NoProfile -ExecutionPolicy Bypass -File "$PLUGINSDIR\remove-minimtspro-bin.ps1" -Target "$INSTDIR\bin"'
    Pop $0
    DetailPrint "Removed $INSTDIR\bin from system PATH (exit code $0)"
FunctionEnd
