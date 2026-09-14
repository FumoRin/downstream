package downloader

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func createServerForFilename(t *testing.T, filename string, payload []byte) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
		w.Header().Set("Accept-Ranges", "bytes")
		http.ServeContent(w, r, filename, time.Now(), bytes.NewReader(payload))
	})
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func TestDownload_CustomDestinationDir(t *testing.T) {
	payload := []byte("custom directory test payload")
	server := createServerForFilename(t, "data.bin", payload)

	customDir := filepath.Join(t.TempDir(), "my-custom-folder")

	opts := DownloadOptions{
		Dir:      customDir,
		Progress: make(chan Progress, 10),
	}

	totalSize, finalPath, err := Download("job-custom-dir", server.URL, opts, context.Background())
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	expectedPath := filepath.Join(customDir, "data.bin")
	if finalPath != expectedPath {
		t.Errorf("expected path %q, got %q", expectedPath, finalPath)
	}

	if totalSize != int64(len(payload)) {
		t.Errorf("expected size %d, got %d", len(payload), totalSize)
	}

	data, err := os.ReadFile(finalPath)
	if err != nil {
		t.Fatalf("failed to read downloaded file at %s: %v", finalPath, err)
	}
	if !bytes.Equal(data, payload) {
		t.Errorf("content mismatch")
	}
}

func TestDownload_ExplicitPathInFilename(t *testing.T) {
	payload := []byte("explicit filename path test payload")
	server := createServerForFilename(t, "ignored_original.bin", payload)

	customDir := filepath.Join(t.TempDir(), "explicit-folder")
	explicitPath := filepath.Join(customDir, "renamed_output.dat")

	opts := DownloadOptions{
		Filename: explicitPath,
		Progress: make(chan Progress, 10),
	}

	_, finalPath, err := Download("job-explicit-path", server.URL, opts, context.Background())
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	if finalPath != explicitPath {
		t.Errorf("expected path %q, got %q", explicitPath, finalPath)
	}

	data, err := os.ReadFile(finalPath)
	if err != nil {
		t.Fatalf("file not found at %s: %v", finalPath, err)
	}
	if !bytes.Equal(data, payload) {
		t.Errorf("content mismatch")
	}
}

func TestDownload_CategoryRouting(t *testing.T) {
	payload := []byte("video category test content")
	server := createServerForFilename(t, "movie.mp4", payload)

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	downloadBaseDir := filepath.Join(tempDir, "Downloads")
	if err := repo.SeedDefaultCategories(downloadBaseDir); err != nil {
		t.Fatalf("failed to seed categories: %v", err)
	}

	opts := DownloadOptions{
		Repo: repo,
		Settings: &Settings{
			DownloadDir: downloadBaseDir,
		},
		Progress: make(chan Progress, 10),
	}

	_, finalPath, err := Download("job-cat-routing", server.URL, opts, context.Background())
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	expectedPath := filepath.Join(downloadBaseDir, "Videos", "movie.mp4")
	if finalPath != expectedPath {
		t.Errorf("expected destination %q, got %q", expectedPath, finalPath)
	}

	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("file was not saved to expected category folder: %s", expectedPath)
	}
}

func TestDownload_CategoryFallbackToDownloadDir(t *testing.T) {
	payload := []byte("unclassified extension test")
	server := createServerForFilename(t, "archive.unknownext", payload)

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	downloadBaseDir := filepath.Join(tempDir, "Downloads")
	if err := repo.SeedDefaultCategories(downloadBaseDir); err != nil {
		t.Fatalf("failed to seed categories: %v", err)
	}

	opts := DownloadOptions{
		Repo: repo,
		Settings: &Settings{
			DownloadDir: downloadBaseDir,
		},
		Progress: make(chan Progress, 10),
	}

	_, finalPath, err := Download("job-cat-fallback", server.URL, opts, context.Background())
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	expectedPath := filepath.Join(downloadBaseDir, "archive.unknownext")
	if finalPath != expectedPath {
		t.Errorf("expected destination %q, got %q", expectedPath, finalPath)
	}

	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("file was not saved in fallback download dir: %s", expectedPath)
	}
}

func TestDownloadManager_WithCustomDirAndCategory(t *testing.T) {
	payload := []byte("manager custom dir test")
	server := createServerForFilename(t, "tune.mp3", payload)

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	downloadBaseDir := filepath.Join(tempDir, "Downloads")
	_ = repo.SeedDefaultCategories(downloadBaseDir)

	settings := &Settings{
		DownloadDir:            downloadBaseDir,
		MaxConcurrencyDownload: 2,
	}

	mgr := NewDownloadManager(repo, settings)
	mgr.StartWorker(1)

	// Test 1: Download to explicit custom dir
	customFolder := filepath.Join(tempDir, "custom-music-folder")
	jobID := "job-mgr-custom-dir"
	mgr.StartDownload(jobID, server.URL, "", customFolder, nil)

	mgr.Close()
	mgr.Wait()

	expectedFile := filepath.Join(customFolder, "tune.mp3")
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("expected file at %s, not found", expectedFile)
	}

	state, err := repo.GetDownload(jobID)
	if err != nil || state == nil {
		t.Fatalf("failed to get state: %v", err)
	}
	if state.Filename != expectedFile {
		t.Errorf("expected state.Filename %q, got %q", expectedFile, state.Filename)
	}
	if state.Category != "Music" {
		t.Errorf("expected state.Category 'Music', got %q", state.Category)
	}
}
