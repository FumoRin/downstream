package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fumorin/gdl-manager/internal/downloader"
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
		gdlDir := filepath.Join(configDir, "gdl")
		_ = os.MkdirAll(gdlDir, 0o755)
		dbPath = filepath.Join(gdlDir, "gdl.db")
	} else {
		dbPath = "gdl.db"
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
