package downloader

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/time/rate"
)

// StartWorker Start a download worker for each job
func (m *DownloadManager) StartWorker(count int) {
	for range count {
		go func() {
			for {
				job, ok := <-m.job
				if !ok {
					return
				}
				m.processJob(job)
			}
		}()
	}
}

// StartDownload Start a Download
func (m *DownloadManager) StartDownload(generatedID, url, filename, dir string, limiters ...*rate.Limiter) {
	var limiter *rate.Limiter
	if len(limiters) > 0 {
		limiter = limiters[0]
	}

	state, err := m.repo.GetDownload(generatedID)
	if err != nil || state == nil {
		state = &DownloadState{
			ID:        generatedID,
			URL:       url,
			Filename:  filename,
			TotalSize: 0,
		}
	}
	state.Status = StateQueue

	if err := m.repo.SaveDownload(state); err != nil {
		fmt.Printf("database error: %v\n", err)
		return
	}

	m.wg.Add(1)

	select {
	case <-m.ctx.Done():
		state.Status = StatePaused
		if err := m.repo.SaveDownload(state); err != nil {
			fmt.Printf("database error: %v\n", err)
		}
		m.wg.Done()
		return

	case m.job <- DownloadJob{ID: generatedID, URL: url, Filename: filename, Dir: dir,Limiter: limiter}:
		return
	}
}

func (m *DownloadManager) StopDownload(id string) {
	m.cancellationsMu.Lock()
	cancel, ok := m.cancellations[id]
	m.cancellationsMu.Unlock()

	if ok {
		cancel()
	}
}

func (m *DownloadManager) RenameDownload(id string, newFilename string) error {
	state, err := m.repo.GetDownload(id)
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("download not found")
	}

	m.StopDownload(id)

	// Handle physical file rename
	oldTmp := state.Filename + "." + id + ".tmp"
	newTmp := newFilename + "." + id + ".tmp"

	if _, err := os.Stat(oldTmp); err == nil {
		if err := os.Rename(oldTmp, newTmp); err != nil {
			return fmt.Errorf("failed to rename temp file: %w", err)
		}
	}

	// If it's completed, we should also rename the final file.
	// Since we don't store the exact final filename (GetUniqueFilename adds suffixes),
	// this part is tricky. But we can try to rename state.Filename if it exists.
	if state.Status == StateCompleted {
		if _, err := os.Stat(state.Filename); err == nil {
			if err := os.Rename(state.Filename, newFilename); err != nil {
				// We ignore this error because GetUniqueFilename might have renamed it to something else
				fmt.Printf("warning: could not rename final file %s: %v\n", state.Filename, err)
			}
		}
	}

	if err := m.repo.UpdateFilename(id, newFilename); err != nil {
		return err
	}

	return nil
}

func (m *DownloadManager) DeleteDownload(id string, deleteFiles bool) error {
	state, err := m.repo.GetDownload(id)
	if err != nil {
		return err
	}
	if state == nil {
		return fmt.Errorf("download not found")
	}

	m.StopDownload(id)

	if deleteFiles {
		// Delete temp file
		tmpFile := state.Filename + "." + id + ".tmp"
		_ = os.Remove(tmpFile)
		// Delete final file
		_ = os.Remove(state.Filename)
	}

	if err := m.repo.DeleteDownload(id); err != nil {
		return err
	}

	return nil
}

func NewDownloadManager(repo DownloadRepository, settings *Settings) *DownloadManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &DownloadManager{
		job:           make(chan DownloadJob, 1000),
		progress:      make(chan Progress),
		repo:          repo,
		cancellations: make(map[string]context.CancelFunc),
		ctx:           ctx,
		cancel:        cancel,
		settings:      settings,
	}
}

func (m *DownloadManager) processJob(job DownloadJob) {
	defer m.wg.Done()
	if m.ctx.Err() != nil {
		state, err := m.repo.GetDownload(job.ID)
		if err != nil {
			return
		}
		state.Status = StatePaused
		if err := m.repo.SaveDownload(state); err != nil {
			fmt.Printf("database error: %v\n", err)
		}
		return
	}

	// If it's downloading
	ctx, cancel := context.WithCancel(m.ctx)
	m.cancellationsMu.Lock()
	m.cancellations[job.ID] = cancel
	m.cancellationsMu.Unlock()

	opts := DownloadOptions{
		URL:      job.URL,
		Filename: job.Filename,
		Dir:      job.Dir,
		Progress: m.progress,
		Repo:     m.repo,
		Settings: m.settings,
		Limiter:  job.Limiter,
	}

	totalSize, filename, downloadErr := Download(job.ID, job.URL, opts, ctx)

	state, err := m.repo.GetDownload(job.ID)
	if err != nil {
		return
	}

	if filename != "" {
		state.Filename = filename
		if m.repo != nil && m.settings != nil {
			if categories, catErr := m.repo.GetCategories(); catErr == nil {
				_, catName := ResolveDestination(filepath.Base(filename), categories, m.settings.DownloadDir)
				state.Category = catName
			}
		}
	}

	if totalSize > 0 {
		state.TotalSize = totalSize
	}

	if downloadErr == nil {
		state.Status = StateCompleted
		if err := m.repo.DeletePart(job.ID); err != nil {
			fmt.Printf("Failed to clean up workers tracking %s: %v\n", state.Filename, err)
		}
	} else if errors.Is(downloadErr, context.Canceled) {
		state.Status = StatePaused
	} else {
		state.Status = StateError

		fmt.Printf("Download Failed for %s: %v\n", job.URL, downloadErr)
	}

	if err := m.repo.SaveDownload(state); err != nil {
		fmt.Printf("database error: %v\n", err)
		return
	}

	m.cancellationsMu.Lock()
	delete(m.cancellations, job.ID)
	m.cancellationsMu.Unlock()

	cancel()
}

func (m *DownloadManager) ScheduleDownload(id, url, customFilename string, ctx context.Context, scheduledAt time.Time, limiters ...*rate.Limiter) (*TargetInfo, error) {
	info, err := m.Probe(ctx, url, customFilename)
	resolvedFilename := customFilename
	var totalSize int64
	if err == nil && info != nil {
		resolvedFilename = info.Filename
		totalSize = info.TotalSize
	} else {
		if resolvedFilename == "" {
			resolvedFilename = FallbackFilenameFromURL(url)
		}
	}

	state := &DownloadState{
		ID: id,
		URL: url,
		Filename: resolvedFilename,
		TotalSize: totalSize,
		Status: StateScheduled,
		ScheduledAt: &scheduledAt,
	}

	if saveErr := m.repo.SaveDownload(state); saveErr != nil {
		return nil, saveErr
	}

	return info, err
}

func (m *DownloadManager) StartScheduler(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-m.ctx.Done():
				return 
			case <-ticker.C:
				m.checkScheduledDownloads()
			}
		}
	}()
}

func (m *DownloadManager) checkScheduledDownloads() {
	due, err := m.repo.GetDueScheduledDownloads(time.Now())
	if err != nil || len(due) == 0 {
		return 
	}

	for _, item := range due {
		m.StartDownload(item.ID, item.URL, item.Filename, "",  nil)
	}
}

func (m *DownloadManager) Probe(ctx context.Context, url, customFilename string) (*TargetInfo, error) {
	opts := DownloadOptions{
		Filename: customFilename,
		Settings: m.settings,
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	return probeServerSupport(ctx, client, url, opts)
}


// === Public API & Lifecycle ===

// Shutdown : Handling SIGTERM by doing a Graceful shutdown
func (m *DownloadManager) Shutdown() {
	m.cancel()
}

// Close : Replacement for closing a m.Job channel
func (m *DownloadManager) Close() {
	close(m.job)
}

// Wait : Replacement for m.Wg.Wait()
func (m *DownloadManager) Wait() {
	m.wg.Wait()
	close(m.progress)
}

// ProgressChan : Passing ProgressChan for Read-Only Access on Progress channel
func (m *DownloadManager) ProgressChan() <- chan Progress {
	return m.progress
}

// GetDownload : Repo access delegation for Get Download
func (m *DownloadManager) GetDownload(id string) (*DownloadState, error) {
	return m.repo.GetDownload(id)
}

// GetIncompleteDownload : Repo access delegation for Get Incomplete Download
func (m *DownloadManager) GetIncompleteDownload() ([]*DownloadState, error) {
	return m.repo.GetIncompleteDownload()
}

// GetAllDownload : Repo access delegation for Get Incomplete Download
func (m *DownloadManager) GetAllDownload() ([]*DownloadState, error) {
	return m.repo.GetAllDownloads()
}
