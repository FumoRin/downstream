package downloader

import (
	"path/filepath"
	"testing"
)

func TestSettingsPersistence(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	// 1. Verify default values
	settings, err := repo.GetAllSettings()
	if err != nil {
		t.Fatalf("failed to get defaults: %v", err)
	}
	if settings.MaxConcurrencyDownload != 3 || settings.PartsPerDownload != 4 {
		t.Errorf("unexpected defaults: %+v", settings)
	}

	// 2. Set new values
	if err := repo.SetSettings("max_concurrent_download", "5"); err != nil {
		t.Fatalf("failed to set setting: %v", err)
	}
	if err := repo.SetSettings("per_part_download", "8"); err != nil {
		t.Fatalf("failed to set setting: %v", err)
	}
	if err := repo.SetSettings("download_dir", "/custom/downloads"); err != nil {
		t.Fatalf("failed to set setting: %v", err)
	}

	// 3. Verify updated values persist
	updated, err := repo.GetAllSettings()
	if err != nil {
		t.Fatalf("failed to get updated settings: %v", err)
	}
	if updated.MaxConcurrencyDownload != 5 {
		t.Errorf("expected MaxConcurrencyDownload=5, got %d", updated.MaxConcurrencyDownload)
	}
	if updated.PartsPerDownload != 8 {
		t.Errorf("expected PartsPerDownload=8, got %d", updated.PartsPerDownload)
	}
	if updated.DownloadDir != "/custom/downloads" {
		t.Errorf("expected DownloadDir='/custom/downloads', got %s", updated.DownloadDir)
	}
}

func TestCategoryDestinationResolution(t *testing.T) {
	categories := []*Category{
		{
			Name:       "Videos",
			FolderPath: "/downloads/Videos",
			Extension:  []string{"mp4", "mkv"},
		},
		{
			Name:       "Documents",
			FolderPath: "/downloads/Documents",
			Extension:  []string{"pdf", "docx"},
		},
	}

	// Test video file matching
	dest, catName := ResolveDestination("movie.mp4", categories, "/downloads")
	if catName != "Videos" || filepath.Dir(dest) != "/downloads/Videos" {
		t.Errorf("expected Videos category, got %s (dest: %s)", catName, dest)
	}

	// Test document file matching
	dest, catName = ResolveDestination("sheet.docx", categories, "/downloads")
	if catName != "Documents" || filepath.Dir(dest) != "/downloads/Documents" {
		t.Errorf("expected Documents category, got %s (dest: %s)", catName, dest)
	}

	// Test unclassified extension fallback
	dest, catName = ResolveDestination("unknown.xyz", categories, "/downloads")
	if catName != "General" || filepath.Dir(dest) != "/downloads" {
		t.Errorf("expected General fallback, got %s (dest: %s)", catName, dest)
	}
}
