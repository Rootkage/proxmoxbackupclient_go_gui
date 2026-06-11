# Proxmox Backup Client - Directory Backup GUI

This directory contains the Fyne-based GUI for the Directory Backup functionality.

## Features

- **Graphical Configuration**: Easy-to-use interface for configuring all backup parameters
- **Configuration Management**: Load and save configuration files (JSON format)
- **Real-time Logging**: View backup progress and logs in real-time
- **SMTP Notifications**: Configure email notifications for backup completion/failure
- **Directory Browser**: Browse and select directories to backup
- **Progress Tracking**: Visual progress bar during backup operations
- **Cross-platform**: Works on Windows, Linux, and macOS

## Screenshots

The GUI includes two main tabs:

1. **Backup Tab**: Configure all backup parameters
   - Proxmox Backup Server connection settings
   - Authentication credentials
   - Backup source directory selection
   - Optional SMTP notification settings
   - Start/Stop backup buttons
   - Progress bar and status display

2. **Logs Tab**: View and manage backup logs
   - Real-time log display
   - Clear logs button
   - Save logs to file

## Requirements

- Go 1.24.4 or later
- Fyne v2.5.3 or later

## Installation

### 1. Install Dependencies

```bash
# Navigate to the directorybackup directory
cd directorybackup

# Get Go modules
go mod tidy
```

### 2. Build the GUI

#### For Windows:
```bash
# Build for Windows
go build -tags gui -o directorybackup_gui.exe .
```

#### For Linux:
```bash
go build -tags gui -o directorybackup_gui .
```

#### For macOS:
```bash
go build -tags gui -o directorybackup_gui .
```

Or use the provided build scripts:
- Windows: `build_gui.bat`
- Linux/macOS: `./build_gui.sh`

## Usage

### Running the GUI

```bash
# Run the GUI application
./directorybackup_gui
```

Or on Windows:
```cmd
directorybackup_gui.exe
```

### Configuration

1. **Server Settings**:
   - **Base URL**: The URL of your Proxmox Backup Server (e.g., `https://192.168.1.10:8007`)
   - **Certificate Fingerprint**: The SSL certificate fingerprint (leave empty for insecure connections)
   - **Auth ID**: Your PBS API token (format: `user@realm!apiid`)
   - **Secret**: Your API token secret
   - **Datastore**: The datastore to backup to
   - **Namespace**: Optional namespace
   - **Backup ID**: Optional backup identifier (defaults to hostname)

2. **Backup Settings**:
   - **Backup Directory**: The directory to backup (use the Browse button to select)
   - **PXAR Output**: Optional path to save PXAR archive for debugging
   - **Use VSS**: Enable Volume Shadow Copy for consistent backups (Windows only)

3. **SMTP Notifications (Optional)**:
   - **Host**: SMTP server hostname
   - **Port**: SMTP server port
   - **Username**: SMTP username
   - **Password**: SMTP password
   - **From**: Sender email address
   - **To**: Recipient email address
   - **Allow Insecure Connection**: Enable for servers without SSL

### Loading/Saving Configuration

- **Load Config**: Click to load a previously saved configuration from a JSON file
- **Save Config**: Click to save the current configuration to a JSON file

### Running a Backup

1. Fill in all required fields
2. Click "Start Backup"
3. Monitor progress in the status bar and logs tab
4. Receive a notification when backup completes

### Viewing Logs

- Switch to the "Logs" tab to view real-time backup logs
- Use "Clear Logs" to clear the log display
- Use "Save Logs" to save logs to a file

## Configuration File Format

The GUI uses the same JSON configuration format as the CLI version:

```json
{
  "baseurl": "https://192.168.1.10:8007",
  "certfingerprint": "ea:7d:06:f9:87:73:a4:72:d0:e8:05:a4:b3:3d:95:d7:0a:26:dd:6d:5c:ca:e6:99:83:e4:11:3b:5f:10:f4:4b",
  "authid": "user@realm!apiid",
  "secret": "your-secret",
  "datastore": "your-datastore",
  "namespace": "optional-namespace",
  "backup-id": "optional-backup-id",
  "backupdir": "C:\\path\\to\\backup",
  "pxarout": "optional-pxar-path",
  "usevss": true,
  "smtp": {
    "host": "smtp.example.com",
    "port": "587",
    "username": "user@example.com",
    "password": "password",
    "insecure": false,
    "mails": [
      {
        "from": "sender@example.com",
        "to": "recipient@example.com"
      }
    ],
    "template": {
      "subject": "Backup {{.Status}}",
      "body": "Backup complete! New: {{.NewChunks}}, Reused: {{.ReusedChunks}}"
    }
  }
}
```

## Troubleshooting

### Common Issues

1. **Fyne dependencies not found**:
   ```bash
   go mod tidy
   ```

2. **Build fails with missing imports**:
   Make sure all dependencies are installed:
   ```bash
   go get fyne.io/fyne/v2
   go get github.com/cornelk/hashmap
   ```

3. **GUI doesn't start**:
   - On Linux, you may need to install additional dependencies:
     ```bash
     # Ubuntu/Debian
     sudo apt-get install libgl1-mesa-dev xorg-dev
     
     # Fedora
     sudo dnf install mesa-libGL-devel libX11-devel
     ```

4. **High DPI display issues**:
   Fyne should handle this automatically, but if you experience issues, try:
   ```bash
   export FYNE_SCALE=1.5  # or 2.0 for higher DPI
   ./directorybackup_gui
   ```

## Architecture

The GUI is built using the [Fyne](https://fyne.io/) framework, which provides native-looking cross-platform GUIs.

### Main Components

- **DirectoryBackupGUI**: Main GUI structure that manages state and UI
- **GUIConfig**: Extended configuration with GUI-specific fields
- **BackupProgress**: Tracks backup progress metrics
- **createMainTab()**: Creates the main configuration tab
- **createLogsTab()**: Creates the logs display tab
- **runBackup()**: Executes the backup process in a background goroutine
- **sendMailNotification()**: Sends email notifications after backup completion

### Threading Model

- The GUI runs in the main thread
- Backup operations run in background goroutines
- UI updates are marshaled back to the main thread using Fyne's notification system
- Progress updates are sent via atomic counters

## Contributing

To contribute to the GUI:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test on multiple platforms (Windows, Linux, macOS)
5. Submit a pull request

### Adding New Features

- Add new widgets to the appropriate tab
- Update the configuration structure if needed
- Ensure proper error handling and user feedback
- Test on all supported platforms

## License

This software is licensed under the GPLv3 license, same as the main project.

## Support

For support and questions, please refer to the main project's README and issue tracker.
