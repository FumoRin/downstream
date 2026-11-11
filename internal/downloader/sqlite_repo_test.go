package downloader

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSQLiteRepository_CRUD(t *testing.T) {
	tempDir := t.TempDir()
	repo, err := NewSQLiteRepository(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	state := &DownloadState{
		ID:        "dl-1",
		URL:       "http://example.com/file.zip",
		Filename:  "file.zip",
		TotalSize: 1048576,
		Status:    StateDownloading,
	}

	// 1. Save and Get
	if err := repo.SaveDownload(state); err != nil {
		t.Fatalf("SaveDownload failed: %v", err)
	}

	got, err := repo.GetDownload("dl-1")
	if err != nil || got == nil {
		t.Fatalf("GetDownload failed: %v", err)
	}
	if got.Filename != "file.zip" || got.TotalSize != 1048576 {
		t.Errorf("unexpected retrieved state: %+v", got)
	}

	// 2. Update Filename
	if err := repo.UpdateFilename("dl-1", "renamed.zip"); err != nil {
		t.Fatalf("UpdateFilename failed: %v", err)
	}
	got, _ = repo.GetDownload("dl-1")
	if got.Filename != "renamed.zip" {
		t.Errorf("expected filename 'renamed.zip', got '%s'", got.Filename)
	}

	// 3. GetAllDownloads
	all, err := repo.GetAllDownloads()
	if err != nil || len(all) != 1 {
		t.Fatalf("expected 1 download, got %d", len(all))
	}

	// 4. DeleteDownload
	if err := repo.DeleteDownload("dl-1"); err != nil {
		t.Fatalf("DeleteDownload failed: %v", err)
	}
	got, _ = repo.GetDownload("dl-1")
	if got != nil {
		t.Errorf("expected nil after delete, got %+v", got)
	}
}

func TestSQLiteRepository_IncompleteDownloads(t *testing.T) {
	tempDir := t.TempDir()
	repo, err := NewSQLiteRepository(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	_ = repo.SaveDownload(&DownloadState{ID: "1", Status: StateCompleted})
	_ = repo.SaveDownload(&DownloadState{ID: "2", Status: StatePaused})
	_ = repo.SaveDownload(&DownloadState{ID: "3", Status: StateQueue})

	incomplete, err := repo.GetIncompleteDownload()
	if err != nil {
		t.Fatalf("failed to query incomplete: %v", err)
	}

	if len(incomplete) != 2 {
		t.Errorf("expected 2 incomplete downloads, got %d", len(incomplete))
	}
}

func TestSQLiteRepository_PartsAndCascadeDelete(t *testing.T) {
	tempDir := t.TempDir()
	repo, err := NewSQLiteRepository(filepath.Join(tempDir, "test.db"))
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	// 1. Create parent download
	_ = repo.SaveDownload(&DownloadState{ID: "dl-parent", Status: StateDownloading})

	// 2. Create parts
	part1 := &PartState{
		ID:          "part-1",
		DownloadID:  "dl-parent",
		StartByte:   0,
		EndByte:     500,
		CurrentByte: 100,
		WorkerID:    "worker-1",
	}
	if err := repo.CreatePart(part1); err != nil {
		t.Fatalf("CreatePart failed: %v", err)
	}

	// 3. Update part progress
	if err := repo.UpdatePartsProgress("part-1", 250); err != nil {
		t.Fatalf("UpdatePartsProgress failed: %v", err)
	}

	parts, _ := repo.GetParts("dl-parent")
	if len(parts) != 1 || parts[0].CurrentByte != 250 {
		t.Errorf("expected current_byte 250, got %+v", parts)
	}

	// 4. Test Cascade Delete: Deleting parent must automatically delete parts!
	_ = repo.DeleteDownload("dl-parent")
	remainingParts, _ := repo.GetParts("dl-parent")
	if len(remainingParts) != 0 {
		t.Errorf("foreign key cascade failed: expected 0 parts, found %d", len(remainingParts))
	}
}

func TestSQLiteRepository_SchemaVersion(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "version_test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	version, err := repo.GetSchemaVersion()
	if err != nil {
		t.Fatalf("GetSchemaVersion failed: %v", err)
	}
	if version != 3 {
		t.Errorf("expected schema version 3, got %d", version)
	}
	_ = repo.Close()

	// Reopening must remain idempotent at version 2 without error
	reopened, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to reopen repo: %v", err)
	}
	reopenedVer, _ := reopened.GetSchemaVersion()
	if reopenedVer != 3 {
		t.Errorf("expected schema version 3 on reopen, got %d", reopenedVer)
	}
}

func TestSQLiteRepository_VersionGuard(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "guard_test.db")

	repo, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	// Simulate a database from a newer version (version 99)
	if _, err := repo.db.Exec("PRAGMA user_version = 99;"); err != nil {
		t.Fatalf("failed to set future version: %v", err)
	}
	_ = repo.Close()

	// Reopening with current binary must not fail or downgrade blindly
	reopened, err := NewSQLiteRepository(dbPath)
	if err != nil {
		t.Fatalf("expected version guard to allow opening in compatibility mode, got error: %v", err)
	}
	v, _ := reopened.GetSchemaVersion()
	if v != 99 {
		t.Errorf("expected version to remain 99, got %d", v)
	}
}

func TestSQLiteRepository_HeadersPersistence(t *testing.T) {
	tempDir := t.TempDir()
	repo, err := NewSQLiteRepository(filepath.Join(tempDir, "headers_test.db"))
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	state := &DownloadState{
		ID:        "auth-dl-1",
		URL:       "https://example.com/protected.zip",
		Filename:  "protected.zip",
		TotalSize: 1024,
		Status:    StateQueue,
		Headers: map[string]string{
			"Cookie":        "session_id=abcdef123456",
			"Authorization": "Bearer token_xyz",
			"User-Agent":    "Downstream/1.0",
		},
	}

	if err := repo.SaveDownload(state); err != nil {
		t.Fatalf("SaveDownload failed: %v", err)
	}

	// 1. Verify via GetDownload
	got, err := repo.GetDownload("auth-dl-1")
	if err != nil || got == nil {
		t.Fatalf("GetDownload failed: %v", err)
	}
	if got.Headers == nil || got.Headers["Cookie"] != "session_id=abcdef123456" {
		t.Errorf("expected Cookie header preserved, got %+v", got.Headers)
	}
	if got.Headers["Authorization"] != "Bearer token_xyz" {
		t.Errorf("expected Authorization header preserved, got %+v", got.Headers)
	}

	// 2. Verify via GetAllDownloads
	all, err := repo.GetAllDownloads()
	if err != nil || len(all) != 1 {
		t.Fatalf("GetAllDownloads failed: %v", err)
	}
	if all[0].Headers == nil || all[0].Headers["User-Agent"] != "Downstream/1.0" {
		t.Errorf("expected User-Agent in GetAllDownloads, got %+v", all[0].Headers)
	}
}

func TestSQLiteRepository_CreatedAtPersistence(t *testing.T) {
	tempDir := t.TempDir()
	repo, err := NewSQLiteRepository(filepath.Join(tempDir, "test_created_at.db"))
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	initialTime := time.Date(2026, 11, 10, 15, 30, 0, 0, time.UTC)
	state := &DownloadState{
		ID:        "test-created-at",
		URL:       "http://example.com/file.zip",
		Filename:  "file.zip",
		CreatedAt: initialTime,
		Status:    StateQueue,
	}

	// 1. Save and verify CreatedAt is retrieved by GetDownload
	if err := repo.SaveDownload(state); err != nil {
		t.Fatalf("SaveDownload failed: %v", err)
	}

	got, err := repo.GetDownload("test-created-at")
	if err != nil || got == nil {
		t.Fatalf("GetDownload failed: %v", err)
	}
	if got.CreatedAt.Unix() != initialTime.Unix() {
		t.Errorf("expected CreatedAt %v, got %v", initialTime.Unix(), got.CreatedAt.Unix())
	}

	// 2. Verify GetAllDownloads retrieves CreatedAt
	all, err := repo.GetAllDownloads()
	if err != nil || len(all) == 0 {
		t.Fatalf("GetAllDownloads failed: %v", err)
	}
	if all[0].CreatedAt.Unix() != initialTime.Unix() {
		t.Errorf("expected all[0].CreatedAt %v, got %v", initialTime.Unix(), all[0].CreatedAt.Unix())
	}

	// 3. Verify updating status does NOT overwrite initial CreatedAt (coalesce behavior)
	state.Status = StateCompleted
	state.CreatedAt = initialTime.Add(2 * time.Hour) // Attempt to change timestamp on update
	if err := repo.SaveDownload(state); err != nil {
		t.Fatalf("SaveDownload update failed: %v", err)
	}

	updated, err := repo.GetDownload("test-created-at")
	if err != nil || updated == nil {
		t.Fatalf("GetDownload after update failed: %v", err)
	}
	if updated.CreatedAt.Unix() != initialTime.Unix() {
		t.Errorf("expected CreatedAt to remain %v, but was overwritten to %v", initialTime.Unix(), updated.CreatedAt.Unix())
	}
}
