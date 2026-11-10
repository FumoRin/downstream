package downloader

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDownloadWithHeaders_AuthenticationAndCookies(t *testing.T) {
	payload := []byte("confidential user content 1234567890")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie := r.Header.Get("Cookie")
		auth := r.Header.Get("Authorization")

		if cookie != "session_id=secret123" || auth != "Bearer token_xyz" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		http.ServeContent(w, r, "secure.bin", time.Now(), bytes.NewReader(payload))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer func() {
		_ = os.Chdir(originalWd)
	}()

	headers := map[string]string{
		"Cookie":        "session_id=secret123",
		"Authorization": "Bearer token_xyz",
		"User-Agent":    "DownstreamTest/1.0",
	}

	// 1. ProbeURL without headers must fail (401 Unauthorized)
	_, err := ProbeURL(context.Background(), server.URL)
	if err == nil {
		t.Error("expected error probing without authentication headers, got nil")
	}

	// 2. ProbeURLWithHeaders with headers must succeed
	info, err := ProbeURLWithHeaders(context.Background(), server.URL, headers)
	if err != nil {
		t.Fatalf("ProbeURLWithHeaders failed: %v", err)
	}
	if info.TotalSize != int64(len(payload)) {
		t.Errorf("expected size %d, got %d", len(payload), info.TotalSize)
	}

	// 3. Download with headers must succeed and write file
	opts := DownloadOptions{
		Filename: "secure.bin",
		Headers:  headers,
		Progress: make(chan Progress, 10),
	}
	totalSize, finalFilename, err := Download("secure-job-1", server.URL, opts, context.Background())
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}
	if totalSize != int64(len(payload)) {
		t.Errorf("expected size %d, got %d", len(payload), totalSize)
	}

	content, err := os.ReadFile(finalFilename)
	if err != nil || !bytes.Equal(content, payload) {
		t.Errorf("file content mismatch, got %s, expected %s", string(content), string(payload))
	}
}

func TestDownloadManagerQueue(t *testing.T) {
	payload := []byte("small file content")
	server := CreateTestRangeServer(t, payload)

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}
	setting, err := repo.GetAllSettings()
	if err != nil {
		t.Fatalf("failed to retrieved settings: %v", err)
	}
	setting.DownloadDir = tempDir

	mgr := NewDownloadManager(repo, setting)
	mgr.StartWorker(2)
	for i := range 4 {
		id := fmt.Sprintf("job-%d", i)
		mgr.StartDownload(id, server.URL, fmt.Sprintf("file-%d.bin", i), "", nil)
	}

	mgr.Close()
	mgr.Wait()

	download, _ := repo.GetAllDownloads()
	for _, d := range download {
		if d.Status != StateCompleted {
			t.Errorf("job %s has status %v, expected StateCompleted", d.ID, d.Status)
		}
	}
}

func TestDownloadManagerStop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1000000")
		w.WriteHeader(http.StatusOK)
		for range 100 {
			_, _ = w.Write(make([]byte, 1000))
			w.(http.Flusher).Flush()
			time.Sleep(10 * time.Millisecond)
		}
	}))
	defer server.Close()

	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}
	setting, err := repo.GetAllSettings()
	if err != nil {
		t.Fatalf("failed to retrieved settings: %v", err)
	}
	setting.DownloadDir = tempDir

	mgr := NewDownloadManager(repo, setting)
	mgr.StartWorker(1)

	JobID := "job-to-stop"
	mgr.StartDownload(JobID, server.URL, "slow.bin", "", nil)

	time.Sleep(30 * time.Millisecond)

	mgr.StopDownload(JobID)
	mgr.Close()
	mgr.Wait()

	state, err := repo.GetDownload(JobID)
	if err != nil {
		t.Fatalf("failed to query download list: %v", err)
	}

	if state.Status != StatePaused {
		t.Errorf("expected statue StatePaused, got %s", state.Status)
	}
}

func TestDownloadManagerDelete(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}
	setting, err := repo.GetAllSettings()
	if err != nil {
		t.Fatalf("failed to retrieved settings: %v", err)
	}

	mgr := NewDownloadManager(repo, setting)

	filepath := filepath.Join(tempDir, "sample.bin")
	_ = os.WriteFile(filepath, []byte("test content"), 0o644)

	jobID := "job-del"
	_ = repo.SaveDownload(&DownloadState{
		ID:       jobID,
		URL:      "https://example.com/sample.bin",
		Filename: filepath,
		Status:   StateCompleted,
	})

	err = mgr.DeleteDownload(jobID, false)
	if err != nil {
		t.Fatalf("failed to delete download record: %v", err)
	}

	state, _ := mgr.GetDownload(jobID)
	if state != nil {
		t.Errorf("expected download record to be deleted from DB is still exist")
	}

	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Errorf("file should be not deleted from disk when deleteFiles=false")
	}

	_ = repo.SaveDownload(&DownloadState{
		ID:       jobID,
		Filename: filepath,
		Status:   StateCompleted,
	})

	err = mgr.DeleteDownload(jobID, true)
	if err != nil {
		t.Fatalf("failed to delete download with files: %v", err)
	}

	if _, err := os.Stat(filepath); !os.IsNotExist(err) {
		t.Errorf("file should be removed from disk when deletedFiles=true")
	}
}
