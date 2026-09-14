package downloader

import (
	"database/sql"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	conn.SetMaxOpenConns(1)

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	if _, err := conn.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, err
	}
	metadataQuery := `
	CREATE TABLE IF NOT EXISTS download_metadata (
		id TEXT PRIMARY KEY,
		url TEXT NOT NULL,
		filename TEXT NOT NULL,
		total_size INTEGER,
		status INTEGER,
		categories TEXT,
		scheduled_at INTEGER
	);
	`

	partQuery := `
	CREATE TABLE IF NOT EXISTS download_part_progress (
		id TEXT PRIMARY KEY,
		download_id TEXT REFERENCES download_metadata(id) ON DELETE CASCADE,
		start_byte INTEGER NOT NULL,
		end_byte INTEGER NOT NULL,
		current_byte INTEGER NOT NULL,
		workers_id TEXT
	);
	`

	settingsQuery := `
	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	)
	`

	categoriesQuery := `
	CREATE TABLE IF NOT EXISTS categories (
		name TEXT PRIMARY KEY,
		folder_path TEXT NOT NULL,
		extension TEXT NOT NULL
	)
	`

	if _, err := conn.Exec(metadataQuery); err != nil {
		return nil, err
	}
	if _, err := conn.Exec(partQuery); err != nil {
		return nil, err
	}
	if _, err := conn.Exec(settingsQuery); err != nil {
		return nil, err
	}
	if _, err := conn.Exec(categoriesQuery); err != nil {
		return nil, err
	}
	return &SQLiteRepository{db: conn}, nil
}

func (r *SQLiteRepository) SaveDownload(state *DownloadState) error {
	var scheduledUnix sql.NullInt64
	if state.ScheduledAt  != nil {
		scheduledUnix = sql.NullInt64{Int64: state.ScheduledAt.Unix(), Valid: true}
	}

	query := `
	INSERT INTO download_metadata (id, url, filename, categories, total_size, status, scheduled_at)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET 
		url = excluded.url,
		filename = excluded.filename,
		categories = excluded.categories,
		total_size = excluded.total_size,
		status = excluded.status,
		scheduled_at = excluded.scheduled_at;
	`
	_, err := r.db.Exec(query, state.ID, state.URL, state.Filename, state.Category, state.TotalSize, state.Status, scheduledUnix)
	if err != nil {
		return err
	}

	return nil
}

func (r *SQLiteRepository) GetDownload(id string) (*DownloadState, error) {
	state := &DownloadState{}
	var status int
	var scheduledUnix sql.NullInt64

	err := r.db.QueryRow("SELECT id, url, filename, categories, total_size, status, scheduled_at FROM download_metadata WHERE id = ?", id).Scan(&state.ID, &state.URL, &state.Filename, &state.Category, &state.TotalSize, &status, &scheduledUnix)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	state.Status = DownloadStatus(status)
	if scheduledUnix.Valid {
		t := time.Unix(scheduledUnix.Int64, 0)
		state.ScheduledAt = &t
	}
	return state, nil
}

func scanDownloadRows(rows *sql.Rows) ([]*DownloadState, error) {
	var states []*DownloadState
	for rows.Next() {
		s := &DownloadState{}
		var status int
		var scheduledUnix sql.NullInt64
		if err := rows.Scan(&s.ID, &s.URL, &s.Filename, &s.Category, &s.TotalSize, &status, &scheduledUnix); err != nil {
			return nil, err
		}
		s.Status = DownloadStatus(status)
		if scheduledUnix.Valid {
			t := time.Unix(scheduledUnix.Int64, 0)
			s.ScheduledAt = &t
		}
		states = append(states, s)
	}
	return states, rows.Err()
}

func (r *SQLiteRepository) GetIncompleteDownload() ([]*DownloadState, error) {
	rows, err := r.db.Query("SELECT id, url, filename, categories, total_size, status, scheduled_at FROM download_metadata WHERE status != ?", StateCompleted)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rows.Close()
	}()

	return scanDownloadRows(rows)
}

func (r *SQLiteRepository) GetAllDownloads() ([]*DownloadState, error) {
	rows, err := r.db.Query("SELECT id, url, filename, categories, total_size, status, scheduled_at FROM download_metadata")
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rows.Close()
	}()

	return scanDownloadRows(rows)
}

func (r *SQLiteRepository) DeleteDownload(id string) error {
	query := `DELETE FROM download_metadata WHERE id = ?`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLiteRepository) UpdateFilename(id string, newFilename string) error {
	query := `UPDATE download_metadata SET filename = ? WHERE id = ?`
	_, err := r.db.Exec(query, newFilename, id)
	if err != nil {
		return err
	}
	return nil
}

func (r *SQLiteRepository) CreatePart(part *PartState) error {
	query := `
	INSERT INTO download_part_progress (id, download_id, start_byte, end_byte, current_byte, workers_id)
	VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query, part.ID, part.DownloadID, part.StartByte, part.EndByte, part.CurrentByte, part.WorkerID)
	if err != nil {
		return err
	}

	return nil
}

func (r *SQLiteRepository) UpdatePartsProgress(partID string, currentByte int64) error {
	query := `
	UPDATE download_part_progress SET current_byte = ? WHERE id = ?
	`
	_, err := r.db.Exec(query, currentByte, partID)
	if err != nil {
		return err
	}

	return nil
}

func (r *SQLiteRepository) GetParts(downloadID string) ([]*PartState, error) {
	rows, err := r.db.Query("SELECT id, download_id, start_byte, end_byte, current_byte, workers_id FROM download_part_progress WHERE download_id = ?", downloadID)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rows.Close()
	}()

	var partState []*PartState
	for rows.Next() {
		currentPart := &PartState{}
		if err := rows.Scan(
			&currentPart.ID,
			&currentPart.DownloadID,
			&currentPart.StartByte,
			&currentPart.EndByte,
			&currentPart.CurrentByte,
			&currentPart.WorkerID,
		); err != nil {
			return nil, err
		}

		partState = append(partState, currentPart)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return partState, nil
}

func (r *SQLiteRepository) DeletePart(downloadID string) error {
	query := `
	DELETE FROM download_part_progress WHERE download_id = ?
	`

	_, err := r.db.Exec(query, downloadID)
	if err != nil {
		return err
	}

	return nil
}

func (r *SQLiteRepository) GetAllSettings() (*Settings, error) {
	home, _ := os.UserHomeDir()
	defaultDownloads := filepath.Join(home, "Downloads")

	settings := &Settings{
		DownloadDir:            defaultDownloads,
		MaxConcurrencyDownload: 3,
		PartsPerDownload:       4,
		MinMultiPartDownload:   5 * 1024 * 1024,
	}

	rows, err := r.db.Query("SELECT key, value FROM settings")
	if err != nil {
		return settings, err
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var key, val string
		if err := rows.Scan(&key, &val); err != nil {
			continue
		}

		switch key {
		case "download_dir":
			settings.DownloadDir = val
		case "max_concurrent_download":
			if n, err := strconv.Atoi(val); err == nil {
				settings.MaxConcurrencyDownload = n
			}
		case "per_part_download":
			if n, err := strconv.Atoi(val); err == nil {
				settings.PartsPerDownload = n
			}
		case "min_multi_part_size":
			if n, err := strconv.ParseInt(val, 10, 64); err == nil && n > 0 {
				settings.MinMultiPartDownload = n
			}
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return settings, nil
}

func (r *SQLiteRepository) SetSettings(key, value string) error {
	query := `
		INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value;
	`
	_, err := r.db.Exec(query, key, value)
	return err
}

func (r *SQLiteRepository) SeedDefaultCategories(defaultDownloadDir string) error {
	defaults := []struct {
		name      string
		folder    string
		extension string
	}{
		{"Videos", filepath.Join(defaultDownloadDir, "Videos"), "mp4,mkv,avi,mov,flv,webm"},
		{"Music", filepath.Join(defaultDownloadDir, "Music"), "mp3,wav,flac,aac,ogg,m4a"},                                                                                                                      
		{"Documents", filepath.Join(defaultDownloadDir, "Documents"), "pdf,doc,docx,xls,xlsx,ppt,pptx,txt,epub"},                                                                                               
		{"Archives", filepath.Join(defaultDownloadDir, "Archives"), "zip,rar,7z,tar,gz,bz2,xz"},                                                                                                                
		{"Programs", filepath.Join(defaultDownloadDir, "Programs"), "exe,msi,dmg,deb,rpm,AppImage,iso"},                                                                                                        
	}

	for _, d := range defaults {
		query := `INSERT OR IGNORE INTO categories (name, folder_path, extension) VALUES (?, ?, ?)`
		if _, err := r.db.Exec(query, d.name, d.folder, d.extension); err != nil {
			return err
		}
	}

	return nil
}

func (r *SQLiteRepository) GetCategories() ([]*Category, error) {
	rows, err := r.db.Query("SELECT name, folder_path, extension FROM categories")
	if err != nil {
		return nil, err
	}
	defer func()  {
		_ = rows.Close()
	}()

	var categories []*Category
	for rows.Next() {
		var name, folder, ext string
		if err := rows.Scan(&name, &folder, &ext); err != nil {
			return nil, err
		}

		var extList []string
		for e := range strings.SplitSeq(ext, ",") {
			trimmed := strings.TrimSpace(e)
			if trimmed != "" {
				extList = append(extList, trimmed)
			}
		}

		categories = append(categories, &Category{
			Name: name,
			FolderPath: folder,
			Extension: extList,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (r *SQLiteRepository) GetDueScheduledDownloads(now time.Time) ([]*DownloadState, error) {
	query := `
	SELECT id, url, filename, categories, total_size, status, scheduled_at
	FROM download_metadata
	WHERE status = ? AND scheduled_at IS NOT NULL AND scheduled_at <= ?
	`

	rows, err := r.db.Query(query, StateScheduled, now.Unix())
	if err != nil {
		return nil, err
	}
	defer func()  {
		_ = rows.Close()
	}()

	return scanDownloadRows(rows)
}
