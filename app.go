package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fumorin/downstream/internal/downloader"
	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx      context.Context
	manager  *downloader.DownloadManager
	repo     downloader.DownloadRepository
	settings downloader.Settings
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
		runtime.LogErrorf(ctx, "failed to initialize database: %v", err)
	}
	a.repo = repo

	settings, _ := repo.GetAllSettings()
	_ = repo.SeedDefaultCategories(settings.DownloadDir)
	a.settings = *settings

	a.manager = downloader.NewDownloadManager(repo, settings)
	a.manager.StartWorker(settings.MaxConcurrencyDownload)
	a.manager.StartScheduler(1 * time.Second)

	go func() {
		for prog := range a.manager.ProgressChan() {
			runtime.EventsEmit(a.ctx, "download:progress", prog)
		}
	}()
}

// Shutdown called by Wails when the desktop window is closed
func (a *App) shutdown(ctx context.Context) {
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
	selected, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
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
		return nil, fmt.Errorf("repository not initialize")
	}
	return a.repo.GetParts(id)
}
