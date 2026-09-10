package downloader

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func CreateTestRangeServer(t *testing.T, payload []byte) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "testFile.bin", time.Now(), bytes.NewReader(payload))
	})
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return server
}

func TestDownloadServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))

	defer server.Close()

	tempDir := t.TempDir()
	_ = os.Chdir(tempDir)

	opts := DownloadOptions{Filename: "missing.bin", Progress: make(chan Progress, 10)}
	_, _, err := Download("job-404", server.URL, opts, context.Background())
	if err == nil {
		t.Errorf("unexpected error for 404 response, got nil")
	}
}

func TestSinglePartDownload(t *testing.T) {
	payload := make([]byte, 100*1024)
	_, _ = rand.Read(payload)

	server := CreateTestRangeServer(t, payload)

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	originalWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	opts := DownloadOptions{
		Filename: "downloaded.bin",
		Repo:     repo,
		Progress: make(chan Progress, 100),
	}

	totalSize, filename, err := Download("job-1", server.URL, opts, context.Background())
	if err != nil {
		t.Fatalf("download failed: %v", err)
	}

	if totalSize != int64(len(payload)) {
		t.Fatalf("expected size %d, got %d", len(payload), totalSize)
	}

	downloadedBytes, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if !bytes.Equal(downloadedBytes, payload) {
		t.Fatalf("downloaded content does not match original payload")
	}
}

func TestMultiPartDownload(t *testing.T) {
	payload := make([]byte, 6*1024*1024)
	_, _ = rand.Read(payload)

	server := CreateTestRangeServer(t, payload)

	tempdir := t.TempDir()
	dbPath := filepath.Join(tempdir, "test.db")
	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	originalWd, _ := os.Getwd()
	_ = os.Chdir(tempdir)
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	opts := DownloadOptions{
		Filename: "downloaded.bin",
		Repo:     repo,
		Progress: make(chan Progress, 100),
	}

	err = repo.SaveDownload(&DownloadState{
		ID:        "job-multipart",
		URL:       server.URL,
		Filename:  "resumable.bin",
		TotalSize: int64(len(payload)),
		Status:    StateDownloading,
	})
	if err != nil {
		t.Fatalf("failed to initialize metadata: %v", err)
	}

	totalSize, filename, err := Download("job-multipart", server.URL, opts, context.Background())
	if err != nil {
		t.Fatalf("download failed: %v", err)
	}

	if totalSize != int64(len(payload)) {
		t.Fatalf("expected size %d, got %d", len(payload), totalSize)
	}

	downloadedBytes, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if !bytes.Equal(downloadedBytes, payload) {
		t.Fatalf("downloaded content does not match original payload")
	}
}

func TestPauseAndResume(t *testing.T) {
	payload := make([]byte, 6*1024*1024) // 6 MB
	_, _ = rand.Read(payload)

	server := CreateTestRangeServer(t, payload)

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	originalWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	err = repo.SaveDownload(&DownloadState{
		ID:        "job-resume",
		URL:       server.URL,
		Filename:  "resumable.bin",
		TotalSize: int64(len(payload)),
		Status:    StateDownloading,
	})
	if err != nil {
		t.Fatalf("failed to initialize metadata: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	opts := DownloadOptions{
		Filename: "resumable.bin",
		Repo:     repo,
		Progress: make(chan Progress, 100),
	}

	_, _, err = Download("job-resume", server.URL, opts, ctx)
	if err == nil {
		t.Log("Download finished before cancel fired; increase payload size if needed")
		return
	}

	resumeOpts := DownloadOptions{
		Filename: "resumable.bin",
		Repo:     repo,
		Progress: make(chan Progress, 100),
	}

	totalSize, finalFilename, err := Download("job-resume", server.URL, resumeOpts, context.Background())
	if err != nil {
		t.Fatalf("Resume failed: %v", err)
	}

	if totalSize != int64(len(payload)) {
		t.Errorf("expected size %d, got %d", len(payload), totalSize)
	}

	downloadedBytes, err := os.ReadFile(finalFilename)
	if err != nil {
		t.Fatalf("failed to read file after resume: %v", err)
	}

	if !bytes.Equal(downloadedBytes, payload) {
		t.Errorf("resumed file corrupted: content does not match original payload")
	}
}

func TestMultiPartFallbackWhenNoRange(t *testing.T) {
	payload := make([]byte, 6*1024*1024)
	_, _ = rand.Read(payload)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Accept-Ranges", "none")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(payload)))
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	tempdir := t.TempDir()
	dbPath := filepath.Join(tempdir, "test.db")
	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	originalWd, _ := os.Getwd()
	_ = os.Chdir(tempdir)
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	opts := DownloadOptions{
		Filename: "downloaded.bin",
		Repo:     repo,
		Progress: make(chan Progress, 100),
	}

	err = repo.SaveDownload(&DownloadState{
		ID:        "job-multipart",
		URL:       server.URL,
		Filename:  "resumable.bin",
		TotalSize: int64(len(payload)),
		Status:    StateDownloading,
	})
	if err != nil {
		t.Fatalf("failed to initialize metadata: %v", err)
	}

	totalSize, filename, err := Download("job-multipart", server.URL, opts, context.Background())
	if err != nil {
		t.Fatalf("download failed: %v", err)
	}

	if totalSize != int64(len(payload)) {
		t.Fatalf("expected size %d, got %d", len(payload), totalSize)
	}

	downloadedBytes, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if !bytes.Equal(downloadedBytes, payload) {
		t.Fatalf("downloaded content does not match original payload")
	}
}

func TestGetUniqueFilename(t *testing.T) {
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "sample.txt")
	_ = os.WriteFile(file1, []byte("hello"), 0o644)

	uniqueName := GetUniqueFilename(file1)
	expected := filepath.Join(tempDir, "sample(1).txt")
	if uniqueName != expected {
		t.Errorf("expected %s, got %s", expected, uniqueName)
	}
}
