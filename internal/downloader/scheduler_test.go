package downloader

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestParseScheduleTime(t *testing.T) {
	now := time.Date(2026, 9, 15, 12, 0, 0, 0, time.UTC)

	// 1. Test --in "30m"
	target, err := ParseScheduleTime("", "30m", now)
	if err != nil || target == nil {
		t.Fatalf("failed to parse --in: %v", err)
	}
	if !target.Equal(now.Add(30 * time.Minute)) {
		t.Errorf("expected %v, got %v", now.Add(30*time.Minute), target)
	}

	// 2. Test --at "14:00" (same day)
	target, err = ParseScheduleTime("14:00", "", now)
	if err != nil || target == nil {
		t.Fatalf("failed to parse --at: %v", err)
	}
	expected := time.Date(2026, 9, 15, 14, 0, 0, 0, time.UTC)
	if !target.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, target)
	}

	// 3. Test --at "10:00" (already passed, should schedule for tomorrow)
	target, err = ParseScheduleTime("10:00", "", now)
	if err != nil || target == nil {
		t.Fatalf("failed to parse --at past: %v", err)
	}
	expectedTomorrow := time.Date(2026, 9, 16, 10, 0, 0, 0, time.UTC)
	if !target.Equal(expectedTomorrow) {
		t.Errorf("expected %v, got %v", expectedTomorrow, target)
	}
}

func TestSchedulerTriggersDownload(t *testing.T) {
	tempDir := t.TempDir()
	repo, _ := NewSQLiteRepository(filepath.Join(tempDir, "test.db"))

	mgr := NewDownloadManager(repo, nil)

	pastTime := time.Now().Add(-1 * time.Second)
	_, _ = mgr.ScheduleDownload("job-sched", "http://example.com/file.bin", "file.bin", context.Background(), pastTime, nil)

	// Run scheduler check
	mgr.checkScheduledDownloads()

	// Verify state transitioned to StateQueue in the job channel
	select {
	case job := <-mgr.job:
		if job.ID != "job-sched" {
			t.Errorf("expected job-sched, got %s", job.ID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("scheduler did not trigger due download")
	}
}

func TestScheduleDownloadProbesMetadata(t *testing.T) {
	payload := []byte("hello scheduled world")
	server := CreateTestRangeServer(t, payload)

	tempDir := t.TempDir()
	repo, _ := NewSQLiteRepository(filepath.Join(tempDir, "test.db"))
	mgr := NewDownloadManager(repo, nil)

	futureTime := time.Now().Add(1 * time.Hour)
	info, err := mgr.ScheduleDownload("job-1", server.URL+"/test-file.bin", "", context.Background(), futureTime, nil)
	if err != nil {
		t.Fatalf("unexpected probe error: %v", err)
	}

	if info.TotalSize != int64(len(payload)) {
		t.Errorf("expected size %d, got %d", len(payload), info.TotalSize)
	}

	// Verify database record has filename and size
	state, err := repo.GetDownload("job-1")
	if err != nil || state == nil {
		t.Fatalf("failed to retrieve state: %v", err)
	}

	if state.TotalSize != int64(len(payload)) {
		t.Errorf("DB TotalSize expected %d, got %d", len(payload), state.TotalSize)
	}
	if state.Filename != "test-file.bin" {
		t.Errorf("DB Filename expected 'test-file.bin', got '%s'", state.Filename)
	}
}

func TestScheduleDownloadWithContentDisposition(t *testing.T) {
	payload := []byte("content disposition test payload")
	expectedFilename := "document-v2.pdf"

	// 1. Create a server that sends Content-Disposition header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", `attachment; filename="`+expectedFilename+`"`)
		w.Header().Set("Content-Length", strconv.Itoa(len(payload)))
		w.Header().Set("Accept-Ranges", "bytes")

		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(payload)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	repo, err := NewSQLiteRepository(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}
	mgr := NewDownloadManager(repo, nil)

	// 2. Use a generic URL path ("/get-file?id=999") so URL fallback would be "get-file"
	targetURL := server.URL + "/get-file?id=999"
	futureTime := time.Now().Add(2 * time.Hour)

	info, err := mgr.ScheduleDownload("job-cd", targetURL, "", context.Background(), futureTime)
	if err != nil {
		t.Fatalf("ScheduleDownload failed: %v", err)
	}

	// 3. Verify probed TargetInfo parsed the Content-Disposition header
	if info.Filename != expectedFilename {
		t.Errorf("expected probed filename %q from Content-Disposition, got %q", expectedFilename, info.Filename)
	}
	if info.TotalSize != int64(len(payload)) {
		t.Errorf("expected probed size %d, got %d", len(payload), info.TotalSize)
	}

	// 4. Verify SQLite record has the header's filename and total size stored
	state, err := repo.GetDownload("job-cd")
	if err != nil || state == nil {
		t.Fatalf("failed to retrieve download from DB: %v", err)
	}

	if state.Filename != expectedFilename {
		t.Errorf("DB filename expected %q, got %q (did it fall back to URL?)", expectedFilename, state.Filename)
	}
	if state.TotalSize != int64(len(payload)) {
		t.Errorf("DB total_size expected %d, got %d", len(payload), state.TotalSize)
	}
	if state.Status != StateScheduled {
		t.Errorf("DB status expected StateScheduled (%v), got %v", StateScheduled, state.Status)
	}
}
