package downloader

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

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

	workingWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer func() {
		_ = os.Chdir(workingWd)
	}()

	mgr := NewDownloadManager(repo, setting)
	mgr.StartWorker(2)
	for i := range 4 {
		id := fmt.Sprintf("job-%d", i)
		mgr.StartDownload(id, server.URL, fmt.Sprintf("file-%d.bin", i))
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

	tempDir := os.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}	
	setting, err := repo.GetAllSettings()
	if err != nil {
		t.Fatalf("failed to retrieved settings: %v", err)
	}

	originalWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)

	defer func() {
		_ = os.Chdir(originalWd)
	}()

	mgr := NewDownloadManager(repo, setting)
	mgr.StartWorker(1)

	JobID := "job-to-stop"
	mgr.StartDownload(JobID, server.URL, "slow.bin")

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
	tempDir := os.TempDir()
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
