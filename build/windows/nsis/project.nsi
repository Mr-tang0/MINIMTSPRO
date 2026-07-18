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

# Default definitions if not provided
!ifndef INFO_PROJECTNAME
    !define INFO_PROJECTNAME "MINIMTSPRO"
!endif
!ifndef INFO_COMPANYNAME
    !define INFO_COMPANYNAME "PIMS"
!endif
!ifndef INFO_PRODUCTNAME
    !define INFO_PRODUCTNAME "MINIMTSPRO"
!endif
!ifndef INFO_PRODUCTVERSION
    !define INFO_PRODUCTVERSION "1.0.0"
!endif
!ifndef INFO_COPYRIGHT
    !define INFO_COPYRIGHT "© 2026, My Company"
!endif
!ifndef PRODUCT_EXECUTABLE
    !define PRODUCT_EXECUTABLE "${INFO_PROJECTNAME}.exe"
!endif
!ifndef UNINST_KEY_NAME
    !define UNINST_KEY_NAME "${INFO_COMPANYNAME}${INFO_PRODUCTNAME}"
!endif
!ifndef REQUEST_EXECUTION_LEVEL
    !define REQUEST_EXECUTION_LEVEL "admin"
!endif
!ifndef WAILS_INSTALL_SCOPE
    !define WAILS_INSTALL_SCOPE "machine"
!endif
!ifndef ARCH
    !define ARCH "amd64"
!endif

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
!include "LogicLib.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
!define MUI_FINISHPAGE_NOAUTOCLOSE
!define MUI_ABORTWARNING

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe"

!if "${WAILS_INSTALL_SCOPE}" == "user"
    InstallDir "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}"
!else
    InstallDir "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
!endif

ShowInstDetails show

Function .onInit
    !insertmacro wails.checkArchitecture
FunctionEnd

; ============================================
; Install Section
; ============================================
Section
    !insertmacro wails.setShellContext
    !insertmacro wails.webview2runtime

    ; Install CP210x driver
    DetailPrint "Installing CP210x driver..."
    InitPluginsDir
    SetOutPath "$PLUGINSDIR\CP210x_VCP_Windows"
    
    !if "${ARCH}" == "amd64"
        File "..\..\..\docs\CP210x_VCP_Windows\CP210xVCPInstaller_x64.exe"
        File "..\..\..\docs\CP210x_VCP_Windows\dpinst.xml"
        File "..\..\..\docs\CP210x_VCP_Windows\slabvcp.cat"
        File "..\..\..\docs\CP210x_VCP_Windows\slabvcp.inf"
        SetOutPath "$PLUGINSDIR\CP210x_VCP_Windows\x64"
        File "..\..\..\docs\CP210x_VCP_Windows\x64\silabenm.sys"
        File "..\..\..\docs\CP210x_VCP_Windows\x64\silabser.sys"
        File "..\..\..\docs\CP210x_VCP_Windows\x64\WdfCoInstaller01009.dll"
        ExecWait '"$PLUGINSDIR\CP210x_VCP_Windows\CP210xVCPInstaller_x64.exe" /S /SE' $0
    !else
        File "..\..\..\docs\CP210x_VCP_Windows\CP210xVCPInstaller_x86.exe"
        File "..\..\..\docs\CP210x_VCP_Windows\dpinst.xml"
        File "..\..\..\docs\CP210x_VCP_Windows\slabvcp.cat"
        File "..\..\..\docs\CP210x_VCP_Windows\slabvcp.inf"
        SetOutPath "$PLUGINSDIR\CP210x_VCP_Windows\x86"
        File "..\..\..\docs\CP210x_VCP_Windows\x86\silabenm.sys"
        File "..\..\..\docs\CP210x_VCP_Windows\x86\silabser.sys"
        File "..\..\..\docs\CP210x_VCP_Windows\x86\WdfCoInstaller01009.dll"
        ExecWait '"$PLUGINSDIR\CP210x_VCP_Windows\CP210xVCPInstaller_x86.exe" /S /SE' $0
    !endif
    
    ${If} $0 != 0
        DetailPrint "Warning: CP210x driver installation may have failed (exit code: $0)"
    ${Else}
        DetailPrint "CP210x driver installed successfully"
    ${EndIf}

    SetOutPath $INSTDIR
    !insertmacro wails.files

    File "..\..\..\update.json"
    File "..\..\..\.env"

    SetOutPath "$INSTDIR\bin"
    File "C:\msys64\ucrt64\bin\*.dll"
    File "C:\opencv\build\bin\*.dll"

    ; Install MVS Runtime once per application lifecycle
    DetailPrint "Installing MVS Runtime..."
    ReadRegStr $1 HKLM "Software\PIMS\MINIMTSPRO" "MVSRuntime_Installed"
    ${If} $1 != "1"
        SetOutPath "$PLUGINSDIR\MVS"
        File "..\..\..\docs\MVS_SDK_V4_8_0_3_MVFG_V2_8_0_3_VC90_Runtime_STD.exe"
        ExecWait '"$PLUGINSDIR\MVS\MVS_SDK_V4_8_0_3_MVFG_V2_8_0_3_VC90_Runtime_STD.exe"' $1
        ${If} $1 == 0
        ${OrIf} $1 == 3010}
            WriteRegStr HKLM "Software\PIMS\MINIMTSPRO" "MVSRuntime_Installed" "1"
            DetailPrint "MVS Runtime installed successfully"
        ${Else}
            DetailPrint "Warning: MVS Runtime installation failed (exit code: $1)"
        ${EndIf}
    ${Else}
        DetailPrint "MVS Runtime already installed, skipping"
    ${EndIf}

    ; Create shortcuts
    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    ; File associations
    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols

    ; Write uninstaller
    !insertmacro wails.writeUninstaller

    ; ============================================
    ; Add program bin to system PATH
    ; ============================================

    DetailPrint "Checking system PATH..."

    ReadRegStr $0 HKLM \
    "SYSTEM\CurrentControlSet\Control\Session Manager\Environment" \
    "Path"


    StrCpy $1 "$INSTDIR\bin"


    InitPluginsDir

    FileOpen $2 "$PLUGINSDIR\add-path.ps1" w

    FileWrite $2 "$$key = 'HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment'$\r$\n"
    FileWrite $2 "$$path = (Get-ItemProperty -LiteralPath $$key -Name Path).Path$\r$\n"
    FileWrite $2 "$$target = '$1'$\r$\n"
    FileWrite $2 "$$target = $$target.TrimEnd('\')$\r$\n"
    FileWrite $2 "$$exists = $$path -split ';' | ForEach-Object { $$_.Trim().TrimEnd('\') } | Where-Object { $$_ -ieq $$target }$\r$\n"
    FileWrite $2 "if ($$exists) {$\r$\n"
    FileWrite $2 "  Write-Host 'Path already exists'$\r$\n"
    FileWrite $2 "} else {$\r$\n"
    FileWrite $2 "  $$newPath = $$path.TrimEnd(';') + ';' + $$target$\r$\n"
    FileWrite $2 "  Set-ItemProperty -LiteralPath $$key -Name Path -Value $$newPath -Type ExpandString$\r$\n"
    FileWrite $2 "  Write-Host 'Path added'$\r$\n"
    FileWrite $2 "}$\r$\n"

    FileClose $2


    nsExec::ExecToLog \
    '"$SYSDIR\WindowsPowerShell\v1.0\powershell.exe" -NoProfile -ExecutionPolicy Bypass -File "$PLUGINSDIR\add-path.ps1"'


    Pop $3


    ${If} $3 == 0
        DetailPrint "PATH checked successfully"
    ${Else}
        DetailPrint "Warning: PATH update failed"
    ${EndIf}


    SendMessage ${HWND_BROADCAST} ${WM_WININICHANGE} 0 "STR:Environment" /TIMEOUT=5000
SectionEnd


; ============================================
; Uninstall Section
; ============================================
Section "uninstall"
    !insertmacro wails.setShellContext

    ; 1. Remove program path from system PATH
    Call un.removeBinFromPath

    ; 2. Stop running processes to avoid file locks
    DetailPrint "Closing application processes..."
    nsExec::ExecToLog '"$SYSDIR\taskkill.exe" /F /IM ${PRODUCT_EXECUTABLE}'
    Sleep 1000

    ; 3. Delete shortcuts
    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"
    Delete "$SMPROGRAMS\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}.lnk"
    RMDir "$SMPROGRAMS\${INFO_COMPANYNAME}"

    ; 4. Delete WebView2 data
    DetailPrint "Deleting WebView2 data..."
    RMDir /r "$AppData\${INFO_PROJECTNAME}"
    RMDir /r "$AppData\${INFO_PRODUCTNAME}"
    RMDir /r "$LOCALAPPDATA\${INFO_PROJECTNAME}"
    RMDir /r "$LOCALAPPDATA\${INFO_PRODUCTNAME}"

    ; 5. Delete program files based on user preference
    ${If} $R0 == "delete"
        DetailPrint "Deleting program directory and all user data..."
        RMDir /r $INSTDIR
    ${Else}
        DetailPrint "Deleting program files, preserving user data..."
        Delete "$INSTDIR\${PRODUCT_EXECUTABLE}"
        Delete "$INSTDIR\*.dll"
        Delete "$INSTDIR\*.exe"
        Delete "$INSTDIR\*.json"
        Delete "$INSTDIR\update.json"
        Delete "$INSTDIR\.env"
        
        RMDir /r "$INSTDIR\bin"
        RMDir "$INSTDIR\resources"
        RMDir "$INSTDIR\locales"
        RMDir $INSTDIR
        
        MessageBox MB_OK|MB_ICONINFORMATION "Application uninstalled. User data preserved at:$\n$INSTDIR$\n(Delete manually for complete removal)"
    ${EndIf}

    ; 6. Clean registry
    DetailPrint "Cleaning registry..."
    DeleteRegKey HKLM "Software\${INFO_COMPANYNAME}\${INFO_PROJECTNAME}"
    DeleteRegKey HKLM "Software\${INFO_PROJECTNAME}"
    DeleteRegKey HKCU "Software\${INFO_COMPANYNAME}\${INFO_PROJECTNAME}"
    DeleteRegKey HKCU "Software\${INFO_PROJECTNAME}"
    DeleteRegKey HKLM "Software\PIMS\MINIMTSPRO"

    ; 7. Remove file associations
    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols

    ; 8. Delete uninstaller
    !insertmacro wails.deleteUninstaller

    ; 9. Notify environment change
    SendMessage ${HWND_BROADCAST} ${WM_WININICHANGE} 0 "STR:Environment" /TIMEOUT=5000

    ; 10. Remove empty install directory if it exists
    ${If} $R0 == "delete"
        RMDir $INSTDIR
    ${EndIf}

    DetailPrint "Uninstall completed successfully!"
SectionEnd

; ============================================
; Uninstall initialization
; ============================================
Function un.onInit
    !insertmacro wails.setShellContext
    
    MessageBox MB_YESNO|MB_ICONQUESTION "Do you want to keep application user data?$\n$\n(Select 'Yes' to keep configuration and log files, 'No' to delete everything)" \
        IDYES noDeleteData IDNO deleteData
    noDeleteData:
        StrCpy $R0 "keep"
        Goto done
    deleteData:
        StrCpy $R0 "delete"
    done:
FunctionEnd

; ============================================
; Remove program bin from system PATH
; ============================================

Function un.removeBinFromPath
    DetailPrint "Removing program path from system PATH..."

    InitPluginsDir

    FileOpen $0 "$PLUGINSDIR\remove-path.ps1" w
    FileWrite $0 "$$key = 'HKLM:\SYSTEM\CurrentControlSet\Control\Session Manager\Environment'$\r$\n"
    FileWrite $0 "$$path = (Get-ItemProperty -LiteralPath $$key -Name Path).Path$\r$\n"
    FileWrite $0 "$$target = '$INSTDIR\bin'$\r$\n"
    FileWrite $0 "$$target = $$target.TrimEnd('\')$\r$\n"
    FileWrite $0 "if ($$path) {$\r$\n"
    FileWrite $0 "  $$newPath = ($$path.Split(';') | Where-Object { $$_.Trim() -and $$_.Trim().TrimEnd('\') -ine $$target }) -join ';'$\r$\n"
    FileWrite $0 "  Set-ItemProperty -LiteralPath $$key -Name Path -Value $$newPath -Type ExpandString$\r$\n"
    FileWrite $0 "  Write-Host 'PATH cleaned'$\r$\n"
    FileWrite $0 "} else {$\r$\n"
    FileWrite $0 "  Write-Host 'PATH empty'$\r$\n"
    FileWrite $0 "}$\r$\n"
    FileClose $0

    nsExec::ExecToLog \
    '"$SYSDIR\WindowsPowerShell\v1.0\powershell.exe" -NoProfile -ExecutionPolicy Bypass -File "$PLUGINSDIR\remove-path.ps1"'

    Pop $1
    ${If} $1 == 0
        DetailPrint "Program PATH removed successfully"
    ${Else}
        DetailPrint "Warning: Failed to remove PATH"
    ${EndIf}

    SendMessage ${HWND_BROADCAST} ${WM_WININICHANGE} 0 "STR:Environment" /TIMEOUT=5000

FunctionEnd