# Implementation Notes for Directory Backup GUI

## Overview

This document provides technical implementation details for the Fyne-based GUI added to the Directory Backup functionality.

## File Structure

```
directorybackup/
├── main.go          # Modified: Added GUI mode detection
├── config.go        # Unchanged: Configuration loading logic
├── go.mod           # Modified: Added Fyne dependency
├── go.sum           # Will be updated by go mod tidy
├── gui.go           # NEW: Main GUI implementation
├── build_gui.sh     # NEW: Build script for Linux/macOS
├── build_gui.bat    # NEW: Build script for Windows
├── GUI_README.md    # NEW: User documentation
└── CHANGES_SUMMARY.md # NEW: Summary of all changes
```

## Key Design Decisions

### 1. Integration Approach

**Decision**: Integrate GUI into the existing main.go rather than creating a separate executable.

**Rationale**:
- Single binary for both CLI and GUI modes
- Users can choose their preferred interface
- No code duplication
- Easier maintenance

**Implementation**:
- Added `-gui`/`--gui` flag detection at the start of `main()`
- When GUI flag is detected, `RunGUI()` is called instead of CLI logic
- The flag is removed from `os.Args` to avoid conflicts with flag parsing

### 2. Framework Selection

**Decision**: Use Fyne v2 (https://fyne.io/)

**Rationale**:
- Native-looking cross-platform GUIs
- Good Go integration
- Active development and community
- Supports Windows, Linux, and macOS
- Good widget library
- Built-in theming support

### 3. Architecture Pattern

**Decision**: Use a struct-based approach with methods for GUI management

**Rationale**:
- Encapsulates GUI state and logic
- Clean separation of concerns
- Easy to extend and maintain
- Follows Go idioms

**Implementation**:
- `DirectoryBackupGUI` struct holds all GUI state
- Methods for creating UI components (`createMainTab`, `createLogsTab`)
- Methods for business logic (`runBackup`, `sendMailNotification`)
- Methods for configuration management (`LoadConfigFromFile`, `SaveConfigToFile`)

### 4. Threading Model

**Decision**: Use goroutines for backup operations with Fyne's notification system for UI updates

**Rationale**:
- Prevents UI freezing during long operations
- Fyne provides thread-safe UI update mechanisms
- Atomic counters for progress tracking
- Clean separation between UI and business logic

**Implementation**:
- Backup runs in a goroutine
- UI updates via `App.SendNotification()`
- Progress tracking via `atomic.Uint64`
- Log updates via shared buffer with periodic polling

### 5. Configuration Management

**Decision**: Reuse existing Config struct and JSON format

**Rationale**:
- Backward compatibility with CLI
- Users can switch between CLI and GUI seamlessly
- No need to maintain separate configuration formats
- Existing validation logic can be reused

**Implementation**:
- `GUIConfig` extends `Config` with GUI-specific fields
- Same JSON marshaling/unmarshaling
- Load/Save functionality integrated into GUI

## Technical Implementation Details

### GUI Initialization

```go
func NewDirectoryBackupGUI() *DirectoryBackupGUI {
    return &DirectoryBackupGUI{
        Config:    &GUIConfig{Config: &Config{UseVSS: true}},
        Progress:  &BackupProgress{},
        LogBuffer: new(string),
        IsRunning: false,
    }
}
```

The GUI is initialized with default values and empty state.

### Main Window Setup

```go
func (gui *DirectoryBackupGUI) Run() {
    gui.App = app.NewWithID("proxmox.directorybackup")
    gui.App.Settings().SetTheme(theme.LightTheme())
    
    gui.Window = gui.App.NewWindow("Proxmox Backup Client - Directory Backup")
    gui.Window.Resize(fyne.NewSize(1024, 768))
    
    // Create tabs
    mainTab := gui.createMainTab()
    logsTab := gui.createLogsTab()
    
    tabs := container.NewAppTabs(
        container.NewTabItem("Backup", mainTab),
        container.NewTabItem("Logs", logsTab),
    )
    
    gui.Window.SetContent(tabs)
    gui.Window.ShowAndRun()
}
```

### Form Creation

The main form is created using Fyne's container and widget system:

- **Layout**: Uses `container.NewVBox()` for vertical stacking
- **Form Fields**: Uses `container.NewHBox()` with labels and input widgets
- **Spacing**: Uses `layout.NewSpacer()` for flexible spacing
- **Scrolling**: Wrapped in `container.NewScroll()` for overflow handling

### Backup Execution

```go
func (gui *DirectoryBackupGUI) runBackup() {
    // Set up client
    insecure := gui.Config.CertFingerprint != ""
    client := &pbscommon.PBSClient{...}
    
    // Set up locking
    L := clientcommon.Locking{}
    lock_ok := L.AcquireProcessLock()
    if !lock_ok {
        // Show error
        return
    }
    defer L.ReleaseProcessLock()
    
    // Connect and run backup
    client.Connect(false, "host")
    begin := time.Now()
    err := backup(client, gui.NewChunk, gui.ReuseChunk, 
                  gui.Config.PxarOut, gui.Config.BackupSourceDir, gui.Config.UseVSS)
    end := time.Now()
    
    // Handle result
    if err != nil {
        gui.AppendLog(fmt.Sprintf("Backup failed: %v", err))
    } else {
        gui.AppendLog(fmt.Sprintf("Backup completed successfully in %s", end.Sub(begin)))
        if gui.Config.SMTP != nil {
            gui.sendMailNotification(begin, end, err)
        }
    }
    client.Finish()
}
```

### Log Management

```go
func (gui *DirectoryBackupGUI) AppendLog(message string) {
    *gui.LogBuffer += message + "\n"
}
```

Logs are appended to a shared buffer. The logs tab displays this buffer and provides buttons to clear or save logs.

## Error Handling

### Validation

- Configuration validation uses existing `Config.valid()` method
- Required fields are checked before starting backup
- Error dialogs are shown for validation failures

### Backup Errors

- Errors during backup are caught and logged
- Fyne notifications are sent for critical errors
- Status label is updated with error information

### UI Errors

- File dialog errors are caught and shown to user
- Configuration load/save errors are logged and shown
- All UI operations have error handling

## Cross-Platform Considerations

### Windows

- VSS support is available via checkbox
- Directory browser works with Windows paths
- High DPI support via Fyne

### Linux

- May require additional system dependencies (libgl1-mesa-dev, xorg-dev)
- Directory browser works with Linux paths
- File permissions are handled appropriately

### macOS

- Native look and feel via Fyne
- Directory browser works with macOS paths
- Retina display support via Fyne

## Performance Considerations

### Memory Usage

- Log buffer grows with each message (could be limited in future)
- Configuration is kept in memory
- Backup operations use existing memory-efficient code

### CPU Usage

- Backup operations run in background
- UI remains responsive
- Progress updates are minimal overhead

### Network Usage

- Same as CLI version (no additional network overhead)
- SMTP notifications add minimal network usage

## Security Considerations

### Credential Storage

- Credentials are stored in plain text in configuration files
- Secret is displayed as password field (masked)
- No credential encryption (same as CLI version)

### File Access

- Configuration files are read/written with user permissions
- Backup directories are accessed with user permissions
- No elevated privileges required

### Network Security

- SSL/TLS is used for PBS connections (unless insecure mode)
- SMTP connections can be configured for SSL/TLS
- Certificate fingerprint verification is supported

## Testing Notes

### Manual Testing

1. **Basic Functionality**
   - Launch GUI with `-gui` flag
   - Verify all form fields are present
   - Verify tab switching works
   - Verify buttons are functional

2. **Configuration**
   - Load a configuration file
   - Verify all fields are populated
   - Modify configuration and save
   - Verify saved file is valid JSON

3. **Backup Execution**
   - Configure with valid PBS server details
   - Start backup
   - Verify progress is shown
   - Verify logs are updated
   - Verify completion notification

4. **Error Handling**
   - Try to start backup with invalid configuration
   - Verify error message is shown
   - Try to load invalid configuration file
   - Verify error message is shown

### Automated Testing

Note: Automated GUI testing is challenging and not currently implemented. Consider:

1. **Unit Tests**: Test non-UI logic (configuration, validation)
2. **Integration Tests**: Test backup functionality (same as CLI)
3. **UI Tests**: Consider Fyne's test utilities for future

## Known Issues and Workarounds

### Issue 1: Stop Button Not Functional

**Current State**: Stop button is disabled

**Workaround**: Close the application to stop backup

**Future Fix**: Implement proper backup cancellation with context

### Issue 2: Progress Bar Not Updating

**Current State**: Progress bar shows/hides but doesn't update percentage

**Workaround**: Monitor logs for progress information

**Future Fix**: Implement progress reporting from backup operations

### Issue 3: Log Updates Not Real-time

**Current State**: Log updates are polled every 500ms

**Workaround**: Acceptable for most use cases

**Future Fix**: Implement event-driven log updates

### Issue 4: High DPI Scaling

**Current State**: Fyne handles most scaling automatically

**Workaround**: Set FYNE_SCALE environment variable if needed

**Future Fix**: Add scaling options to GUI settings

## Future Enhancements

### Short-term (Next Release)

1. **Progress Reporting**: Add progress percentage to progress bar
2. **Stop Functionality**: Implement proper backup cancellation
3. **Configuration Profiles**: Save and load multiple profiles
4. **Recent Files**: Remember recently used configuration files

### Medium-term (Future Release)

1. **Scheduled Backups**: Add scheduling within GUI
2. **Backup History**: Show history of previous backups
3. **Restore Functionality**: Add restore capabilities
4. **Dark Mode**: Add theme switching option
5. **Advanced Settings**: Add more configuration options

### Long-term (Future Considerations)

1. **Localization**: Add support for multiple languages
2. **Dashboard**: Show backup statistics and charts
3. **Multi-tab Backups**: Run multiple backups simultaneously
4. **Backup Verification**: Add verification options
5. **Encryption**: Add encryption support (if implemented in core)

## Migration Guide

### For CLI Users

The GUI is completely optional. Existing CLI usage remains unchanged:

```bash
# CLI usage (unchanged)
./directorybackup -baseurl "..." -authid "..." -secret "..." -backupdir "..." -datastore "..."
```

### For New Users

The GUI provides an easier way to get started:

```bash
# Launch GUI
./directorybackup -gui
```

Then fill in the form and click "Start Backup".

### For Configuration File Users

Existing configuration files work with the GUI:

1. Launch GUI
2. Click "Load Config"
3. Select your existing configuration file
4. All settings will be loaded

## Troubleshooting Guide

### GUI Doesn't Start

**Symptoms**: Application exits immediately or shows error

**Possible Causes**:
1. Missing system dependencies (Linux)
2. Fyne initialization failure
3. Permission issues

**Solutions**:
1. Install required system packages
2. Check error messages in console
3. Run with debug logging if available

### GUI Starts but Backup Fails

**Symptoms**: GUI starts but backup doesn't run or fails

**Possible Causes**:
1. Invalid configuration
2. Connection issues
3. Permission issues

**Solutions**:
1. Check all required fields are filled
2. Verify PBS server is accessible
3. Check logs tab for error messages

### GUI is Slow or Unresponsive

**Symptoms**: GUI is slow to respond

**Possible Causes**:
1. Large log buffer
2. Many UI updates
3. System resource constraints

**Solutions**:
1. Clear logs regularly
2. Close and reopen GUI
3. Check system resources

## Conclusion

The Fyne-based GUI provides a user-friendly interface for the Directory Backup functionality while maintaining full backward compatibility with the existing CLI. The implementation follows Go best practices and Fyne conventions, providing a solid foundation for future enhancements.
