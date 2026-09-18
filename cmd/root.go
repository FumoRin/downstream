package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fumorin/downstream/internal/downloader"
	"github.com/spf13/cobra"
)

type App struct {
	Manager *downloader.DownloadManager
}

var app App

var rootCmd = &cobra.Command{
	Use:   "downstream",
	Short: "Download file from the internet",
	Long:  "A CLI Tool for downloading file from the internet",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
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
			return fmt.Errorf("failed to initialized downloader database: %w", err)
		}

		settings, _ := repo.GetAllSettings()
		_ = repo.SeedDefaultCategories(settings.DownloadDir)
		app.Manager = downloader.NewDownloadManager(repo, settings)
		app.Manager.StartWorker(settings.MaxConcurrencyDownload)
		app.Manager.StartScheduler(1 * time.Second)

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}
