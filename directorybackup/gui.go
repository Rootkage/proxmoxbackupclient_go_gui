package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// GUIConfig extends Config with GUI-specific fields
type GUIConfig struct {
	*Config
	ConfigFilePath string `json:"-"`
}

// BackupProgress tracks the backup progress
type BackupProgress struct {
	NewChunks    uint64
	ReusedChunks uint64
	Status       string
	Error        error
	StartTime    time.Time
	EndTime      time.Time
}

// DirectoryBackupGUI is the main GUI structure
type DirectoryBackupGUI struct {
	App          fyne.App
	Window       fyne.Window
	Config       *GUIConfig
	Progress     *BackupProgress
	LogBuffer    *string
	IsRunning    bool
	NewChunk     *atomic.Uint64
	ReuseChunk   *atomic.Uint64
}

// NewDirectoryBackupGUI creates a new GUI instance
func NewDirectoryBackupGUI() *DirectoryBackupGUI {
	return &DirectoryBackupGUI{
		Config:    &GUIConfig{Config: &Config{UseVSS: true}},
		Progress:  &BackupProgress{},
		LogBuffer: new(string),
		IsRunning: false,
	}
}

// LoadConfigFromFile loads configuration from a JSON file
func (gui *DirectoryBackupGUI) LoadConfigFromFile(path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	
	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return err
	}
	
	gui.Config = &GUIConfig{
		Config:       &config,
		ConfigFilePath: path,
	}
	
	return nil
}

// SaveConfigToFile saves configuration to a JSON file
func (gui *DirectoryBackupGUI) SaveConfigToFile(path string) error {
	configBytes, err := json.MarshalIndent(gui.Config.Config, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(path, configBytes, 0644)
}

// AppendLog appends a message to the log buffer
func (gui *DirectoryBackupGUI) AppendLog(message string) {
	*gui.LogBuffer += message + "\n"
}

// Run starts the Fyne application
func (gui *DirectoryBackupGUI) Run() {
	gui.App = app.NewWithID("proxmox.directorybackup")
	gui.App.Settings().SetTheme(theme.LightTheme())
	
	gui.Window = gui.App.NewWindow("Proxmox Backup Client - Directory Backup")
	gui.Window.Resize(fyne.NewSize(1024, 768))
	
	// Create the main UI
	mainTab := gui.createMainTab()
	logsTab := gui.createLogsTab()
	
	tabs := container.NewAppTabs(
		container.NewTabItem("Backup", mainTab),
		container.NewTabItem("Logs", logsTab),
	)
	
	gui.Window.SetContent(tabs)
	gui.Window.ShowAndRun()
}

// createMainTab creates the main backup configuration tab
func (gui *DirectoryBackupGUI) createMainTab() fyne.CanvasObject {
	// Create form widgets
	baseURLEntry := widget.NewEntry()
	baseURLEntry.SetPlaceHolder("https://192.168.1.10:8007")
	
	certFingerprintEntry := widget.NewEntry()
	certFingerprintEntry.SetPlaceHolder("ea:7d:06:f9:87:73:a4:72:d0:e8:05:a4:b3:3d:95:d7:0a:26:dd:6d:5c:ca:e6:99:83:e4:11:3b:5f:10:f4:4b")
	
	authIDEntry := widget.NewEntry()
	authIDEntry.SetPlaceHolder("user@realm!apiid")
	
	secretEntry := widget.NewPasswordEntry()
	secretEntry.SetPlaceHolder("API Secret")
	
	datastoreEntry := widget.NewEntry()
	datastoreEntry.SetPlaceHolder("datastore")
	
	namespaceEntry := widget.NewEntry()
	namespaceEntry.SetPlaceHolder("namespace (optional)")
	
	backupIDEntry := widget.NewEntry()
	backupIDEntry.SetPlaceHolder("backup-id (optional)")
	
	backupDirEntry := widget.NewEntry()
	backupDirEntry.SetPlaceHolder("C:\\path\\to\\backup")
	
	pxarOutEntry := widget.NewEntry()
	pxarOutEntry.SetPlaceHolder("pxarout (optional)")
	
	useVSSCheck := widget.NewCheck("Use VSS (Volume Shadow Copy)", func(checked bool) {
		gui.Config.UseVSS = checked
	})
	useVSSCheck.SetChecked(true)
	
	// SMTP Configuration (collapsible)
	smtpLabel := widget.NewLabel("SMTP Configuration (Optional)")
	smtpLabel.TextStyle = fyne.TextStyle{Bold: true}
	
	mailHostEntry := widget.NewEntry()
	mailHostEntry.SetPlaceHolder("SMTP Host")
	
	mailPortEntry := widget.NewEntry()
	mailPortEntry.SetPlaceHolder("SMTP Port")
	
	mailUsernameEntry := widget.NewEntry()
	mailUsernameEntry.SetPlaceHolder("SMTP Username")
	
	mailPasswordEntry := widget.NewPasswordEntry()
	mailPasswordEntry.SetPlaceHolder("SMTP Password")
	
	mailFromEntry := widget.NewEntry()
	mailFromEntry.SetPlaceHolder("From Email")
	
	mailToEntry := widget.NewEntry()
	mailToEntry.SetPlaceHolder("To Email")
	
	mailInsecureCheck := widget.NewCheck("Allow Insecure Connection", func(checked bool) {})
	
	// Status bar
	statusLabel := widget.NewLabel("Ready")
	statusLabel.TextStyle = fyne.TextStyle{Italic: true}
	
	// Progress bar
	progressBar := widget.NewProgressBar()
	progressBar.Hide()
	
	// Action buttons
	startButton := widget.NewButton("Start Backup", func() {
		if gui.IsRunning {
			return
		}
		
		// Validate configuration
		gui.Config.BaseURL = baseURLEntry.Text
		gui.Config.CertFingerprint = certFingerprintEntry.Text
		gui.Config.AuthID = authIDEntry.Text
		gui.Config.Secret = secretEntry.Text
		gui.Config.Datastore = datastoreEntry.Text
		gui.Config.Namespace = namespaceEntry.Text
		gui.Config.BackupID = backupIDEntry.Text
		gui.Config.BackupSourceDir = backupDirEntry.Text
		gui.Config.PxarOut = pxarOutEntry.Text
		gui.Config.UseVSS = useVSSCheck.Checked
		
		// SMTP Configuration
		if mailHostEntry.Text != "" || mailPortEntry.Text != "" || mailUsernameEntry.Text != "" || mailPasswordEntry.Text != "" {
			if gui.Config.SMTP == nil {
				gui.Config.SMTP = &SMTPConfig{}
			}
			gui.Config.SMTP.Host = mailHostEntry.Text
			gui.Config.SMTP.Port = mailPortEntry.Text
			gui.Config.SMTP.Username = mailUsernameEntry.Text
			gui.Config.SMTP.Password = mailPasswordEntry.Text
			gui.Config.SMTP.Insecure = mailInsecureCheck.Checked
			
			if mailFromEntry.Text != "" || mailToEntry.Text != "" {
				if len(gui.Config.SMTP.Mails) == 0 {
					gui.Config.SMTP.Mails = append(gui.Config.SMTP.Mails, MailSendConfig{})
				}
				gui.Config.SMTP.Mails[0].From = mailFromEntry.Text
				gui.Config.SMTP.Mails[0].To = mailToEntry.Text
			}
		}
		
		// Validate
		if !gui.Config.valid() {
			dialog.ShowError(fmt.Errorf("Please fill in all required fields"), gui.Window)
			return
		}
		
		// Save config if path is set
		if gui.Config.ConfigFilePath != "" {
			if err := gui.SaveConfigToFile(gui.Config.ConfigFilePath); err != nil {
				gui.AppendLog(fmt.Sprintf("Error saving config: %v", err))
			}
		}
		
		// Start backup in goroutine
		gui.IsRunning = true
		gui.NewChunk = new(atomic.Uint64)
		gui.ReuseChunk = new(atomic.Uint64)
		
		startButton.Disable()
		progressBar.Show()
		progressBar.SetValue(0)
		statusLabel.SetText("Starting backup...")
		gui.AppendLog("Starting backup...")
		
		go func() {
			defer func() {
				gui.IsRunning = false
				startButton.Enable()
				progressBar.Hide()
			}()
			
			// Run the backup
			gui.runBackup()
			
			// Update UI on completion
			gui.App.SendNotification(&fyne.Notification{
				Title:   "Backup Complete",
				Content: fmt.Sprintf("New: %d, Reused: %d", gui.NewChunk.Load(), gui.ReuseChunk.Load()),
			})
			
			statusLabel.SetText(fmt.Sprintf("Backup complete - New: %d, Reused: %d", 
				gui.NewChunk.Load(), gui.ReuseChunk.Load()))
			gui.AppendLog(fmt.Sprintf("Backup complete - New: %d, Reused: %d", 
				gui.NewChunk.Load(), gui.ReuseChunk.Load()))
		}()
	})
	
	stopButton := widget.NewButton("Stop", func() {
		// TODO: Implement stop functionality
		gui.AppendLog("Stop requested (not yet implemented)")
	})
	stopButton.Disable()
	
	// Config file buttons
	loadConfigButton := widget.NewButton("Load Config", func() {
		dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, gui.Window)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()
			
			data, err := io.ReadAll(reader)
			if err != nil {
				dialog.ShowError(err, gui.Window)
				return
			}
			
			var config Config
			if err := json.Unmarshal(data, &config); err != nil {
				dialog.ShowError(err, gui.Window)
				return
			}
			
			// Update UI with loaded config
			baseURLEntry.SetText(config.BaseURL)
			certFingerprintEntry.SetText(config.CertFingerprint)
			authIDEntry.SetText(config.AuthID)
			secretEntry.SetText(config.Secret)
			datastoreEntry.SetText(config.Datastore)
			namespaceEntry.SetText(config.Namespace)
			backupIDEntry.SetText(config.BackupID)
			backupDirEntry.SetText(config.BackupSourceDir)
			pxarOutEntry.SetText(config.PxarOut)
			useVSSCheck.SetChecked(config.UseVSS)
			
			// Update SMTP fields if present
			if config.SMTP != nil {
				mailHostEntry.SetText(config.SMTP.Host)
				mailPortEntry.SetText(config.SMTP.Port)
				mailUsernameEntry.SetText(config.SMTP.Username)
				mailPasswordEntry.SetText(config.SMTP.Password)
				mailInsecureCheck.SetChecked(config.SMTP.Insecure)
				if len(config.SMTP.Mails) > 0 {
					mailFromEntry.SetText(config.SMTP.Mails[0].From)
					mailToEntry.SetText(config.SMTP.Mails[0].To)
				}
			}
			
			gui.Config = &GUIConfig{
				Config:       &config,
				ConfigFilePath: reader.URI().Path(),
			}
			
			gui.AppendLog(fmt.Sprintf("Configuration loaded from: %s", reader.URI().Path()))
		}, gui.Window)
	})
	
	saveConfigButton := widget.NewButton("Save Config", func() {
		// Update config from UI
		gui.Config.BaseURL = baseURLEntry.Text
		gui.Config.CertFingerprint = certFingerprintEntry.Text
		gui.Config.AuthID = authIDEntry.Text
		gui.Config.Secret = secretEntry.Text
		gui.Config.Datastore = datastoreEntry.Text
		gui.Config.Namespace = namespaceEntry.Text
		gui.Config.BackupID = backupIDEntry.Text
		gui.Config.BackupSourceDir = backupDirEntry.Text
		gui.Config.PxarOut = pxarOutEntry.Text
		gui.Config.UseVSS = useVSSCheck.Checked
		
		// SMTP Configuration
		if mailHostEntry.Text != "" || mailPortEntry.Text != "" || mailUsernameEntry.Text != "" || mailPasswordEntry.Text != "" {
			if gui.Config.SMTP == nil {
				gui.Config.SMTP = &SMTPConfig{}
			}
			gui.Config.SMTP.Host = mailHostEntry.Text
			gui.Config.SMTP.Port = mailPortEntry.Text
			gui.Config.SMTP.Username = mailUsernameEntry.Text
			gui.Config.SMTP.Password = mailPasswordEntry.Text
			gui.Config.SMTP.Insecure = mailInsecureCheck.Checked
			
			if mailFromEntry.Text != "" || mailToEntry.Text != "" {
				if len(gui.Config.SMTP.Mails) == 0 {
					gui.Config.SMTP.Mails = append(gui.Config.SMTP.Mails, MailSendConfig{})
				}
				gui.Config.SMTP.Mails[0].From = mailFromEntry.Text
				gui.Config.SMTP.Mails[0].To = mailToEntry.Text
			}
		}
		
		// Show save dialog
		dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, gui.Window)
				return
			}
			if writer == nil {
				return
			}
			defer writer.Close()
			
			configBytes, err := json.MarshalIndent(gui.Config.Config, "", "  ")
			if err != nil {
				dialog.ShowError(err, gui.Window)
				return
			}
			
			_, err = writer.Write(configBytes)
			if err != nil {
				dialog.ShowError(err, gui.Window)
				return
			}
			
			gui.Config.ConfigFilePath = writer.URI().Path()
			gui.AppendLog(fmt.Sprintf("Configuration saved to: %s", writer.URI().Path()))
		}, gui.Window)
	})
	
	// Button row
	buttonRow := container.NewHBox(
		layout.NewSpacer(),
		startButton,
		stopButton,
		loadConfigButton,
		saveConfigButton,
		layout.NewSpacer(),
	)
	
	// SMTP form (collapsible section)
	smtpForm := container.NewVBox(
		smtpLabel,
		widget.NewSeparator(),
		container.NewHBox(
			widget.NewLabel("Host:"),
			layout.NewSpacer(),
			mailHostEntry,
		),
		container.NewHBox(
			widget.NewLabel("Port:"),
			layout.NewSpacer(),
			mailPortEntry,
		),
		container.NewHBox(
			widget.NewLabel("Username:"),
			layout.NewSpacer(),
			mailUsernameEntry,
		),
		container.NewHBox(
			widget.NewLabel("Password:"),
			layout.NewSpacer(),
			mailPasswordEntry,
		),
		container.NewHBox(
			widget.NewLabel("From:"),
			layout.NewSpacer(),
			mailFromEntry,
		),
		container.NewHBox(
			widget.NewLabel("To:"),
			layout.NewSpacer(),
			mailToEntry,
		),
		mailInsecureCheck,
	)
	
	// Main form
	mainForm := container.NewVBox(
		widget.NewLabel("Proxmox Backup Server Configuration"),
		widget.NewLabel("All fields except Namespace, Backup ID, and PXAR Out are required"),
		widget.NewSeparator(),
		
		container.NewHBox(
			widget.NewLabel("Base URL:"),
			layout.NewSpacer(),
			baseURLEntry,
		),
		container.NewHBox(
			widget.NewLabel("Certificate Fingerprint:"),
			layout.NewSpacer(),
			certFingerprintEntry,
		),
		container.NewHBox(
			widget.NewLabel("Auth ID:"),
			layout.NewSpacer(),
			authIDEntry,
		),
		container.NewHBox(
			widget.NewLabel("Secret:"),
			layout.NewSpacer(),
			secretEntry,
		),
		container.NewHBox(
			widget.NewLabel("Datastore:"),
			layout.NewSpacer(),
			datastoreEntry,
		),
		container.NewHBox(
			widget.NewLabel("Namespace:"),
			layout.NewSpacer(),
			namespaceEntry,
		),
		container.NewHBox(
			widget.NewLabel("Backup ID:"),
			layout.NewSpacer(),
			backupIDEntry,
		),
		container.NewHBox(
			widget.NewLabel("Backup Directory:"),
			layout.NewSpacer(),
			backupDirEntry,
		),
		container.NewHBox(
			widget.NewLabel("PXAR Output:"),
			layout.NewSpacer(),
			pxarOutEntry,
		),
		useVSSCheck,
		widget.NewSeparator(),
		
		// SMTP Section
		smtpForm,
		widget.NewSeparator(),
		
		// Directory browser button
		container.NewHBox(
			layout.NewSpacer(),
			widget.NewButton("Browse...", func() {
				dialog.ShowFolderOpen(func(list fyne.ListableURI, err error) {
					if err != nil {
						dialog.ShowError(err, gui.Window)
						return
					}
					if list == nil {
						return
					}
					backupDirEntry.SetText(list.String())
				}, gui.Window)
			}),
			layout.NewSpacer(),
		),
		
		widget.NewSeparator(),
		statusLabel,
		progressBar,
		buttonRow,
	)
	
	// Create a scroll container for the form
	scroll := container.NewScroll(mainForm)
	scroll.SetMinSize(fyne.NewSize(800, 600))
	
	return scroll
}

// createLogsTab creates the logs tab
func (gui *DirectoryBackupGUI) createLogsTab() fyne.CanvasObject {
	logText := widget.NewMultiLineEntry()
	logText.SetPlaceHolder("Logs will appear here...")
	logText.Disable()
	
	// Update log text from buffer
	updateLogText := func() {
		logText.SetText(*gui.LogBuffer)
		// Auto-scroll to bottom
		logText.CursorRow = len(logText.Text)
	}
	
	// Create a button to clear logs
	clearButton := widget.NewButton("Clear Logs", func() {
		*gui.LogBuffer = ""
		updateLogText()
	})
	
	// Create a button to save logs
	saveButton := widget.NewButton("Save Logs", func() {
		dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, gui.Window)
				return
			}
			if writer == nil {
				return
			}
			defer writer.Close()
			
			_, err = writer.Write([]byte(*gui.LogBuffer))
			if err != nil {
				dialog.ShowError(err, gui.Window)
			}
		}, gui.Window)
	})
	
	buttonRow := container.NewHBox(
		layout.NewSpacer(),
		clearButton,
		saveButton,
		layout.NewSpacer(),
	)
	
	// Update log text periodically
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		
		for range ticker.C {
			if *gui.LogBuffer != logText.Text {
				gui.App.SendNotification(&fyne.Notification{
					Title:   "Logs Updated",
					Content: "New log messages available",
				})
			}
		}
	}()
	
	return container.NewBorder(
		buttonRow,
		nil,
		nil,
		nil,
		container.NewScroll(logText),
	)
}

// runBackup runs the actual backup process
func (gui *DirectoryBackupGUI) runBackup() {
	// Set up the client
	insecure := gui.Config.CertFingerprint != ""
	
	client := &pbscommon.PBSClient{
		BaseURL:         gui.Config.BaseURL,
		CertFingerPrint: gui.Config.CertFingerprint,
		AuthID:          gui.Config.AuthID,
		Secret:          gui.Config.Secret,
		Datastore:       gui.Config.Datastore,
		Namespace:       gui.Config.Namespace,
		Insecure:        insecure,
		Manifest: pbscommon.BackupManifest{
			BackupID: gui.Config.BackupID,
		},
	}
	
	gui.AppendLog(fmt.Sprintf("Starting backup of %s", gui.Config.BackupSourceDir))
	
	// Set up locking
	L := clientcommon.Locking{}
	lock_ok := L.AcquireProcessLock()
	if !lock_ok {
		gui.AppendLog("Error: Backup jobs need to run exclusively, please wait until the previous job has finished")
		gui.App.SendNotification(&fyne.Notification{
			Title:   "Error",
			Content: "Backup jobs need to run exclusively, please wait until the previous job has finished",
		})
		return
	}
	defer L.ReleaseProcessLock()
	
	// Connect client
	client.Connect(false, "host")
	
	// Run the backup
	begin := time.Now()
	err := backup(client, gui.NewChunk, gui.ReuseChunk, gui.Config.PxarOut, gui.Config.BackupSourceDir, gui.Config.UseVSS)
	end := time.Now()
	
	if err != nil {
		gui.AppendLog(fmt.Sprintf("Backup failed: %v", err))
		gui.App.SendNotification(&fyne.Notification{
			Title:   "Backup Failed",
			Content: err.Error(),
		})
	} else {
		gui.AppendLog(fmt.Sprintf("Backup completed successfully in %s", end.Sub(begin)))
		
		// Send mail notification if configured
		if gui.Config.SMTP != nil {
			gui.sendMailNotification(begin, end, err)
		}
	}
	
	// Finish the client
	client.Finish()
}

// sendMailNotification sends email notification about backup status
func (gui *DirectoryBackupGUI) sendMailNotification(start, end time.Time, err error) {
	if gui.Config.SMTP == nil {
		return
	}
	
	hostname, _ := os.Hostname()
	
	mailCtx := clientcommon.MailCtx{
		NewChunks:    gui.NewChunk.Load(),
		ReusedChunks: gui.ReuseChunk.Load(),
		Error:        err,
		Hostname:     hostname,
		Datastore:    gui.Config.Datastore,
		StartTime:    start,
		EndTime:      end,
	}
	
	mailBodyTemplate := defaultMailBodyTemplate
	if gui.Config.SMTP.Template != nil && gui.Config.SMTP.Template.Body != "" {
		mailBodyTemplate = gui.Config.SMTP.Template.Body
	}
	
	msg, err := mailCtx.BuildStr(mailBodyTemplate)
	if err != nil {
		gui.AppendLog(fmt.Sprintf("Cannot use custom mail body: %v", err))
		msg, err = mailCtx.BuildStr(defaultMailBodyTemplate)
		if err != nil {
			gui.AppendLog(fmt.Sprintf("Cannot use default mail body: %v", err))
			return
		}
	}
	
	mailSubjectTemplate := defaultMailSubjectTemplate
	if gui.Config.SMTP.Template != nil && gui.Config.SMTP.Template.Subject != "" {
		mailSubjectTemplate = gui.Config.SMTP.Template.Subject
	}
	
	subject, err := mailCtx.BuildStr(mailSubjectTemplate)
	if err != nil {
		gui.AppendLog(fmt.Sprintf("Cannot use custom mail subject: %v", err))
		msg, err = mailCtx.BuildStr(defaultMailSubjectTemplate)
		if err != nil {
			gui.AppendLog(fmt.Sprintf("Cannot use default mail subject: %v", err))
			return
		}
	}
	
	client, err := clientcommon.SetupMailClient(
		gui.Config.SMTP.Host,
		gui.Config.SMTP.Port,
		gui.Config.SMTP.Username,
		gui.Config.SMTP.Password,
		gui.Config.SMTP.Insecure,
	)
	if err != nil {
		gui.AppendLog(fmt.Sprintf("Cannot connect to mail server: %v", err))
		return
	}
	defer client.Quit()
	
	for _, ccc := range gui.Config.SMTP.Mails {
		err = clientcommon.SendMail(ccc.From, ccc.To, subject, msg, client)
		if err != nil {
			gui.AppendLog(fmt.Sprintf("Cannot send email: %v", err))
		} else {
			gui.AppendLog(fmt.Sprintf("Email notification sent to %s", ccc.To))
		}
	}
}

// RunGUI starts the Fyne GUI application
func RunGUI() {
	gui := NewDirectoryBackupGUI()
	gui.Run()
}
