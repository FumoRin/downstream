package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fumorin/downstream/internal/downloader"
	"github.com/google/uuid"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx           context.Context
	manager       *downloader.DownloadManager
	repo          downloader.DownloadRepository
	settings      downloader.Settings
	ipcListener   net.Listener
	ipcSocketPath string
}

type IPCMessage struct {
	Action   string `json:"action"`
	URL      string `json:"url"`
	Filename string `json:"filename,omitempty"`
	Referrer string `json:"referrer,omitempty"`
	Key      string `json:"key,omitempty"`
	Value    string `json:"value,omitempty"`
}

type IPCResponse struct {
	Status   string               `json:"status"`
	ID       string               `json:"id,omitempty"`
	Message  string               `json:"message,omitempty"`
	Settings *downloader.Settings `json:"settings,omitempty"`
}

func (a *App) emitEvent(name string, data ...interface{}) {
	if a.ctx != nil && a.ctx.Value("events") != nil {
		wailsRuntime.EventsEmit(a.ctx, name, data...)
	}
}

func (a *App) showWindow() {
	if a.ctx != nil && a.ctx.Value("window") != nil {
		wailsRuntime.WindowShow(a.ctx)
		wailsRuntime.WindowUnminimise(a.ctx)
	}
}

func (a *App) getIPCSocketPath() string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("downstream-%d.sock", os.Getuid()))
}

func (a *App) startIPCServer() {
	socketPath := a.getIPCSocketPath()
	a.ipcSocketPath = socketPath

	if conn, err := net.DialTimeout("unix", socketPath, 100*time.Millisecond); err == nil {
		_ = conn.Close()
		wailsRuntime.LogWarningf(a.ctx, "another Downstream IPC is already listening on %s", socketPath)
		return
	} else {
		_ = os.Remove(socketPath)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		wailsRuntime.LogErrorf(a.ctx, "failed to start IPC listener: %v", err)
		return
	}
	a.ipcListener = listener
	_ = os.Chmod(socketPath, 0o600)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}

			go a.handleIPCConnection(conn)
		}
	}()
}

func (a *App) handleIPCConnection(conn net.Conn) {
	defer func() {
		_ = conn.Close()
	}()

	var msg IPCMessage
	if err := json.NewDecoder(conn).Decode(&msg); err != nil {
		_ = json.NewEncoder(conn).Encode(IPCResponse{
			Status:  "error",
			Message: fmt.Sprintf("invalid JSON payload: %v", err),
		})
		return
	}

	switch msg.Action {
	case "ping":
		_ = json.NewEncoder(conn).Encode(IPCResponse{Status: "ok"})

	case "get_settings":
		var settings *downloader.Settings
		if a.repo != nil {
			settings, _ = a.repo.GetAllSettings()
		}
		_ = json.NewEncoder(conn).Encode(IPCResponse{
			Status:   "ok",
			Settings: settings,
		})

	case "update_settings":
		if a.repo == nil {
			_ = json.NewEncoder(conn).Encode(IPCResponse{
				Status:  "error",
				Message: "repository not initialized",
			})
			return
		}
		if msg.Key != "" && msg.Value != "" {
			if err := a.repo.SetSettings(msg.Key, msg.Value); err != nil {
				_ = json.NewEncoder(conn).Encode(IPCResponse{
					Status:  "error",
					Message: err.Error(),
				})
				return
			}
		}
		_ = json.NewEncoder(conn).Encode(IPCResponse{Status: "ok"})

	case "open_app":
		if a.ctx != nil {
			a.showWindow()
		}
		_ = json.NewEncoder(conn).Encode(IPCResponse{Status: "ok"})

	case "download":
		if msg.URL == "" {
			_ = json.NewEncoder(conn).Encode(IPCResponse{
				Status:  "error",
				Message: "URL cannot be empty",
			})
			return
		}

		action := "prompt"
		if a.repo != nil {
			if s, err := a.repo.GetAllSettings(); err == nil && s.BrowserDownloadAction != "" {
				action = s.BrowserDownloadAction
			}
		}

		if action == "prompt" {
			// Bring Downstream window into view
			if a.ctx != nil {
				a.showWindow()
			}

			// Emit event for frontend to open AddDownload confirmation modal
			a.emitEvent("download:external_prompt", map[string]string{
				"url":      msg.URL,
				"filename": msg.Filename,
				"referrer": msg.Referrer,
			})

			_ = json.NewEncoder(conn).Encode(IPCResponse{
				Status:  "ok",
				Message: "prompted",
			})
			return
		}

		if action == "auto" {
			autoForeground := false
			if a.repo != nil {
				if s, err := a.repo.GetAllSettings(); err == nil {
					if s.BrowserAutoForeground {
						autoForeground = true
					}
				}
			}

			if autoForeground && a.ctx != nil {
				a.showWindow()
			}
		}

		// Otherwise: "auto" mode (start silently)
		id := uuid.NewString()
		a.manager.StartDownload(id, msg.URL, msg.Filename, "", nil)

		// Notify Wails frontend so the UI refreshes and pops a toast
		a.emitEvent("download:external_added", map[string]string{
			"id":       id,
			"url":      msg.URL,
			"filename": msg.Filename,
		})

		_ = json.NewEncoder(conn).Encode(IPCResponse{
			Status: "ok",
			ID:     id,
		})

	default:
		_ = json.NewEncoder(conn).Encode(IPCResponse{
			Status:  "error",
			Message: fmt.Sprintf("unknown action: %s", msg.Action),
		})
	}
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	configDir, err := os.UserConfigDir()
	var dbPath string
	if err == nil {
		downstreamDir := filepath.Join(configDir, "downstream")
		_ = os.MkdirAll(downstreamDir, 0o755)
		dbPath = filepath.Join(downstreamDir, "downstream_download.db")
	} else {
		dbPath = "downstream_download.db"
	}

	repo, err := downloader.NewSQLiteRepository(dbPath)
	if err != nil {
		wailsRuntime.LogErrorf(ctx, "failed to initialize database: %v", err)
	}
	a.repo = repo

	settings, _ := repo.GetAllSettings()
	_ = repo.SeedDefaultCategories(settings.DownloadDir)
	a.settings = *settings

	a.manager = downloader.NewDownloadManager(repo, settings)
	a.manager.StartWorker(settings.MaxConcurrencyDownload)
	a.manager.StartScheduler(1 * time.Second)
	a.startIPCServer()

	go func() {
		for prog := range a.manager.ProgressChan() {
			wailsRuntime.EventsEmit(a.ctx, "download:progress", prog)
		}
	}()
}

// Shutdown called by Wails when the desktop window is closed
func (a *App) shutdown(ctx context.Context) {
	if a.ipcListener != nil {
		_ = a.ipcListener.Close()
	}
	if a.ipcSocketPath != "" {
		_ = os.Remove(a.ipcSocketPath)
	}
	if a.manager != nil {
		a.manager.Shutdown()
	}
}

// === Exported Go Methods to Wails ===

func (a *App) ProbeURL(url string) (*downloader.TargetInfo, error) {
	if url == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return downloader.ProbeURL(ctx, url)
}

func (a *App) StartDownload(url, customFilename, customDir string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}
	id := uuid.NewString()
	a.manager.StartDownload(id, url, customFilename, customDir, nil)
	return id, nil
}

func (a *App) StopDownload(id string) {
	a.manager.StopDownload(id)
}

func (a *App) ResumeDownload(id string) error {
	state, err := a.manager.GetDownload(id)
	if err != nil || state == nil {
		return fmt.Errorf("download not found")
	}

	a.manager.StartDownload(id, state.URL, state.Filename, "", nil)
	return nil
}

func (a *App) DeleteDownload(id string, deletedFiles bool) error {
	return a.manager.DeleteDownload(id, deletedFiles)
}

func (a *App) GetAllDownloads() ([]*downloader.DownloadState, error) {
	return a.manager.GetAllDownload()
}

func (a *App) GetSettings() (*downloader.Settings, error) {
	return a.repo.GetAllSettings()
}

func (a *App) UpdateSettings(settings downloader.Settings) error {
	if a.repo == nil {
		return fmt.Errorf("repository not initialize")
	}

	if settings.DownloadDir != "" {
		if err := a.repo.SetSettings("download_dir", settings.DownloadDir); err != nil {
			return err
		}
	}

	if settings.MaxConcurrencyDownload != 0 {
		if err := a.repo.SetSettings("max_concurrent_download", strconv.Itoa(settings.MaxConcurrencyDownload)); err != nil {
			return err
		}
	}

	if settings.PartsPerDownload != 0 {
		if err := a.repo.SetSettings("per_part_download", strconv.Itoa(settings.PartsPerDownload)); err != nil {
			return err
		}
	}

	if settings.MinMultiPartDownload != 0 {
		if err := a.repo.SetSettings("min_multi_part_size", strconv.FormatInt(settings.MinMultiPartDownload, 10)); err != nil {
			return err
		}
	}

	if settings.SpeedLimit >= 0 {
		if err := a.repo.SetSettings("speed_limit", strconv.FormatInt(settings.SpeedLimit, 10)); err != nil {
			return err
		}
	}

	if settings.BrowserDownloadAction != "" {
		if err := a.repo.SetSettings("browser_download_action", settings.BrowserDownloadAction); err != nil {
			return err
		}
	}

	if err := a.repo.SetSettings("browser_auto_foreground", strconv.FormatBool(settings.BrowserAutoForeground)); err != nil {
		return err
	}

	a.settings = settings
	if a.manager != nil {
		a.manager.UpdateSettings(&settings)
	}
	return nil
}

func (a *App) SelectDirectory() (string, error) {
	defaultDir := a.settings.DownloadDir
	if defaultDir == "" {
		home, _ := os.UserHomeDir()
		defaultDir = filepath.Join(home, "Downloads")
	}
	selected, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		DefaultDirectory: defaultDir,
		Title:            "Select Download Directory",
	})
	if err != nil {
		return "", err
	}
	return selected, nil
}

func (a *App) GetDiskStorage() (*downloader.DiskStorage, error) {
	dir := a.settings.DownloadDir
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, "Downloads")
	}

	checkDir := dir
	for checkDir != "" && checkDir != "." {
		if _, err := os.Stat(checkDir); err == nil {
			break
		}
		parent := filepath.Dir(checkDir)
		if parent == checkDir {
			break
		}
		checkDir = parent
	}
	if checkDir == "" {
		checkDir = "."
	}

	free, total, err := getDiskSpace(checkDir)
	if err != nil {
		return nil, err
	}
	return &downloader.DiskStorage{
		Free:  free,
		Total: total,
	}, nil
}

func (a *App) GetDownloadParts(id string) ([]*downloader.PartState, error) {
	if a.repo == nil {
		return nil, fmt.Errorf("Repository not initialize")
	}
	return a.repo.GetParts(id)
}

func (a *App) OpenFile(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	cleanPath := filepath.Clean(path)
	if _, err := os.Stat(cleanPath); err != nil {
		return fmt.Errorf("file not found on disk: %w", err)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", cleanPath)
	case "darwin":
		cmd = exec.Command("open", cleanPath)
	default:
		cmd = exec.Command("xdg-open", cleanPath)
	}

	return cmd.Start()
}

func expandPath(path string) string {
	if path == "" {
		return ""
	}

	if path == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
	} else if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return filepath.Clean(path)
}

func (a *App) ShowInFolder(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	cleanPath := expandPath(path)
	dir := cleanPath
	if stat, err := os.Stat(cleanPath); err != nil || !stat.IsDir() {
		dir = filepath.Dir(cleanPath)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", "/select,", cleanPath)
	case "darwin":
		cmd = exec.Command("open", "-R", cleanPath)
	default: // linux
		cmd = exec.Command("xdg-open", dir)
	}
	return cmd.Start()
}

func (a *App) ScheduleDownload(url, customFilename, customDir string, scheduleTimeUnix int64) (string, error) {
	if url == "" {
		return "", fmt.Errorf("url cannot be empty")
	}

	if scheduleTimeUnix <= 0 {
		return "", fmt.Errorf("scheduled time must be a valid time in the future")
	}

	scheduledAt := time.Unix(scheduleTimeUnix, 0)
	if scheduledAt.Before(time.Now()) {
		return "", fmt.Errorf("scheduled time must be in the future")
	}

	id := uuid.NewString()
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	_, err := a.manager.ScheduleDownload(id, url, customFilename, ctx, scheduledAt)
	if err != nil {
		return "", err
	}

	return id, nil
}
