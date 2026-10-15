package main

import (
	"os"
	"testing"

	"github.com/fumorin/downstream/internal/downloader"
)

func TestGetDiskStorage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "downstream_disk_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func ()  {
		_ = os.RemoveAll(tempDir)
	}()

	app := &App{
		settings: downloader.Settings{
			DownloadDir: tempDir,
		},
	}

	storage, err := app.GetDiskStorage()
	if err != nil {
		t.Fatalf("GetDiskStorage failed: %v", err)
	}

	if storage.Total == 0 {
		t.Errorf("expected Total > 0, got %d", storage.Total)
	}
	if storage.Free == 0 {
		t.Errorf("expected Free > 0, got %d", storage.Free)
	}
	if storage.Free > storage.Total {
		t.Errorf("expected Free <= Total, got Free=%d, Total=%d", storage.Free, storage.Total)
	}
}
