# Summary of Changes for Directory Backup GUI

This document summarizes all the changes made to add a Fyne-based GUI to the Directory Backup functionality.

## New Files Created

### 1. `gui.go`
- **Purpose**: Main GUI implementation using Fyne framework
- **Features**:
  - `DirectoryBackupGUI` struct: Main GUI structure managing state and UI
  - `GUIConfig` struct: Extended configuration with GUI-specific fields
  - `BackupProgress` struct: Tracks backup progress metrics
  - `NewDirectoryBackupGUI()`: Creates a new GUI instance
  - `LoadConfigFromFile()`: Loads configuration from JSON file
  - `SaveConfigToFile()`: Saves configuration to JSON file
  - `AppendLog()`: Appends messages to the log buffer
  - `Run()`: Starts the Fyne application
  - `createMainTab()`: Creates the main backup configuration tab
  - `createLogsTab()`: Creates the logs display tab
  - `runBackup()`: Executes the backup process in a background goroutine
  - `sendMailNotification()`: Sends email notifications after backup completion
  - `RunGUI()`: Entry point for the GUI application

### 2. `GUI_README.md`
- **Purpose**: Comprehensive documentation for the GUI
- **Contents**:
  - Features overview
  - Screenshots description
  - Requirements
  - Installation instructions
  - Usage guide
  - Configuration file format
  - Troubleshooting
  - Architecture overview
  - Contributing guidelines
  - License information

### 3. `build_gui.sh`
- **Purpose**: Build script for Linux/macOS
- **Usage**: `./build_gui.sh`

### 4. `build_gui.bat`
- **Purpose**: Build script for Windows
- **Usage**: `build_gui.bat`

## Modified Files

### 1. `go.mod`
- **Change**: Added Fyne v2 dependency
- **Added**: `fyne.io/fyne/v2 v2.5.3`

### 2. `main.go`
- **Change**: Added GUI mode detection and entry point
- **Added**: 
  - GUI mode detection at the start of `main()`
  - Logic to remove `-gui`/`--gui` flags from `os.Args`
  - Call to `RunGUI()` when GUI mode is detected

### 3. `config.go`
- **Change**: No functional changes, but the file was reviewed for consistency

## GUI Features

### Main Tab (Backup Configuration)
1. **Server Configuration**
   - Base URL input field
   - Certificate Fingerprint input field
   - Auth ID input field
   - Secret password field
   - Datastore input field
   - Namespace input field (optional)
   - Backup ID input field (optional)

2. **Backup Settings**
   - Backup Directory input field with Browse button
   - PXAR Output input field (optional)
   - Use VSS checkbox (Windows only)

3. **SMTP Configuration (Optional)**
   - Host, Port, Username, Password fields
   - From and To email fields
   - Allow Insecure Connection checkbox

4. **Action Buttons**
   - Start Backup
   - Stop (placeholder, not yet implemented)
   - Load Config
   - Save Config

5. **Status Display**
   - Status label
   - Progress bar

### Logs Tab
1. **Log Display**
   - Multi-line text area showing real-time logs
   - Auto-scroll to bottom

2. **Log Management**
   - Clear Logs button
   - Save Logs button

## Usage

### Starting the GUI

#### Method 1: Using the `-gui` flag
```bash
# CLI mode (default)
./directorybackup [flags]

# GUI mode
./directorybackup -gui
# or
./directorybackup --gui
```

#### Method 2: Building a dedicated GUI executable
```bash
# Build with GUI tag
go build -tags gui -o directorybackup_gui .

# Run the GUI
./directorybackup_gui
```

### Configuration

All configuration is done through the GUI interface:

1. Fill in the Proxmox Backup Server connection details
2. Enter authentication credentials
3. Select the directory to backup
4. Optionally configure SMTP notifications
5. Click "Start Backup"

### Loading/Saving Configuration

- Use "Load Config" to load a previously saved JSON configuration file
- Use "Save Config" to save the current configuration to a JSON file

## Technical Details

### Architecture

- **Framework**: Fyne v2 (https://fyne.io/)
- **Threading**: Background goroutines for backup operations
- **UI Updates**: Fyne's notification system for thread-safe UI updates
- **Progress Tracking**: Atomic counters for new/reused chunks

### Dependencies

- `fyne.io/fyne/v2`: Main GUI framework
- All existing dependencies from the CLI version are retained

### Cross-Platform Support

- **Windows**: Full support, including VSS integration
- **Linux**: Full support, may require additional system dependencies
- **macOS**: Full support

### System Dependencies (Linux)

For Linux, the following system packages may be required:

```bash
# Ubuntu/Debian
sudo apt-get install libgl1-mesa-dev xorg-dev

# Fedora
sudo dnf install mesa-libGL-devel libX11-devel
```

## Build Instructions

### Prerequisites

1. Go 1.24.4 or later
2. Git (for cloning the repository)

### Steps

```bash
# Clone the repository (if not already done)
git clone https://github.com/Rootkage/proxmoxbackupclient_go_gui.git
cd proxmoxbackupclient_go_gui/directorybackup

# Get dependencies
go mod tidy

# Build for current platform
go build -o directorybackup .

# Run with GUI
./directorybackup -gui

# Or build a dedicated GUI executable
go build -tags gui -o directorybackup_gui .
./directorybackup_gui
```

## Known Limitations

1. **Stop Button**: The stop button is currently a placeholder and not fully implemented
2. **Progress Bar**: The progress bar shows/hides but doesn't update with actual progress percentage
3. **Real-time Log Updates**: Log updates are polled every 500ms rather than being event-driven
4. **High DPI**: While Fyne handles high DPI displays, some manual scaling may be needed on very high DPI screens

## Future Enhancements

Potential improvements for future versions:

1. **Real-time Progress**: Update progress bar with actual backup progress
2. **Stop Functionality**: Implement proper backup cancellation
3. **Configuration Profiles**: Save multiple configuration profiles
4. **Scheduled Backups**: Add scheduling functionality within the GUI
5. **Backup History**: Show history of previous backups
6. **Restore Functionality**: Add restore capabilities
7. **Dark Mode**: Add theme switching
8. **Localization**: Add support for multiple languages
9. **Advanced Settings**: Add more advanced configuration options
10. **Dashboard**: Show backup statistics and charts

## Compatibility

- **Backward Compatible**: The CLI functionality remains unchanged
- **Configuration Compatible**: Uses the same JSON configuration format as the CLI
- **API Compatible**: Uses the same PBS client API as the CLI version

## Testing

The GUI has been designed to work with the existing backup functionality. To test:

1. Build the application
2. Run with `-gui` flag
3. Configure with valid PBS server details
4. Start a backup and verify it completes successfully
5. Test loading/saving configurations
6. Test SMTP notifications (if configured)

## License

All new code is licensed under the same GPLv3 license as the main project.
