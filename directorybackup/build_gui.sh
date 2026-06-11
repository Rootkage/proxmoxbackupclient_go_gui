#!/bin/bash

# Build script for the Directory Backup GUI

# Set the build tag for GUI
export GOOS=windows
export GOARCH=amd64

# Build the GUI version
echo "Building Directory Backup GUI..."
go build -tags gui -o directorybackup_gui.exe .

if [ $? -eq 0 ]; then
    echo "Build successful!"
    echo "GUI executable: directorybackup_gui.exe"
else
    echo "Build failed!"
    exit 1
fi
