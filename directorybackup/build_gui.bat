@echo off
REM Build script for the Directory Backup GUI on Windows

set GOOS=windows
set GOARCH=amd64

echo Building Directory Backup GUI...
go build -tags gui -o directorybackup_gui.exe .

if %ERRORLEVEL% EQU 0 (
    echo Build successful!
    echo GUI executable: directorybackup_gui.exe
) else (
    echo Build failed!
    exit /b 1
)
