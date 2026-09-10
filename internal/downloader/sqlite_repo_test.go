package downloader

import (
	"path/filepath"
	"testing"
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
	repo, _ := NewSQLiteRepository(filepath.Join(tempDir, "test.db"))

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
	repo, _ := NewSQLiteRepository(filepath.Join(tempDir, "test.db"))

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
