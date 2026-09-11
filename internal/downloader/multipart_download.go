package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

func multipartDownload(id string, url string, info *TargetInfo, opts DownloadOptions, ctx context.Context) (int64, string, error) {
	concurrency := 4
	if opts.Settings != nil && opts.Settings.PartsPerDownload > 0 {
		concurrency = opts.Settings.PartsPerDownload
	}
	tmpFilename := info.Filename + "." + id + ".tmp"
	totalSize := info.TotalSize

	file, err := os.OpenFile(tmpFilename, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	if err := file.Truncate(totalSize); err != nil {
		return 0, "", fmt.Errorf("failed to truncate file: %w", err)
	}

	parts, err := getOsPartHelper(id, totalSize, concurrency, opts.Repo)
	if err != nil {
		return 0, "", fmt.Errorf("failed to initialize parts: %w", err)
	}

	var totalDownload atomic.Int64
	for _, p := range parts {
		totalDownload.Add(p.CurrentByte)
	}

	progresTickerCtx, stopTicker := context.WithCancel(ctx)
	defer stopTicker()
	go reportMultiPartProgress(progresTickerCtx, info.Filename, totalSize, &totalDownload, opts.Progress)

	var wg sync.WaitGroup
	errChan := make(chan error, len(parts))
	client := &http.Client{}

	workerCtx, cancelWorkers := context.WithCancel(ctx)
	defer cancelWorkers()

	for _, parts := range parts {
		wg.Add(1)

		go func(p *PartState) {
			defer wg.Done()

			if err := downloadParts(workerCtx, client, url, file, p, opts.Repo, &totalDownload); err != nil {
				select {
				case errChan <- err:
				default:
				}
				cancelWorkers()
			}
		}(parts)
	}

	wg.Wait()
	close(errChan)

	if err := <-errChan; err != nil {
		return totalSize, info.Filename, err
	}

	if err := file.Sync(); err != nil {
		return totalSize, info.Filename, fmt.Errorf("failed to sync file: %w", err)
	}

	_ = file.Close()

	finalFilename := GetUniqueFilename(info.Filename)
	if err := os.Rename(tmpFilename, finalFilename); err != nil {
		return totalSize, info.Filename, fmt.Errorf("failed to rename final name: %w", err)
	}

	if opts.Repo != nil {
		_ = opts.Repo.DeletePart(id)
	}

	if opts.Progress != nil {
		select {
		case opts.Progress <- Progress{
			Filename: finalFilename,
			Percentage: 100.0,
			CurrentSize: totalSize,
			TotalSize: totalSize,
			Speed: 0,
			ETA: 0,
		}:
		default:
		}
	}

	return totalSize, finalFilename, nil
}

func getOsPartHelper(downloadID string, totalSize int64, numParts int, repo DownloadRepository) ([]*PartState, error) {
	if repo != nil {
		existing, err := repo.GetParts(downloadID)
		if err == nil && len(existing) > 0 {
			return existing, nil
		}
	}

	chunkSize := totalSize / int64(numParts)
	var parts []*PartState

	for i := range numParts {
		start := int64(i) * chunkSize
		end := start + chunkSize - 1

		if i == numParts-1 {
			end = totalSize - 1
		}

		part := &PartState{
			ID:          uuid.NewString(),
			DownloadID:  downloadID,
			StartByte:   start,
			EndByte:     end,
			CurrentByte: 0,
			WorkerID:    fmt.Sprintf("worker-%d", i+1),
		}

		if repo != nil {
			if err := repo.CreatePart(part); err != nil {
				return nil, err
			}
		}
		parts = append(parts, part)
	}

	return parts, nil
}

func reportMultiPartProgress(ctx context.Context, filename string, totalSize int64, totalDownload *atomic.Int64, progressChan chan<- Progress) {
	if progressChan == nil {
		return
	}

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	startTime := time.Now()
	startBytes := totalDownload.Load()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			current := totalDownload.Load()
			elapsed := time.Since(startTime).Seconds()
			speed := 0.0
			if elapsed > 0 {
				speed = float64(current-startBytes) / elapsed
			}

			var percentage float64
			var eta time.Duration
			if totalSize > 0 {
				percentage = (float64(current) / float64(totalSize)) * 100
				if speed > 0 {
					remaining := totalSize - current
					eta = time.Duration(float64(remaining)/speed) * time.Second
				}
			}

			update := Progress{
				Filename:    filename,
				Percentage:  percentage,
				CurrentSize: current,
				TotalSize:   totalSize,
				Speed:       speed,
				ETA:         eta,
			}

			select {
			case progressChan <- update:
			default:
			}
		}
	}
}

func downloadParts(
	ctx context.Context,
	client *http.Client,
	url string,
	file *os.File,
	part *PartState,
	repo DownloadRepository,
	totalDownloaded *atomic.Int64,
) error {
	if part.StartByte+part.CurrentByte > part.EndByte {
		return nil
	}

	reqStart := part.StartByte + part.CurrentByte
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", reqStart, part.EndByte))
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("unexpected status: %d for byte range %d-%d (expected 206 Partial Content)", resp.StatusCode, reqStart, part.EndByte)
	}

	buf := make([]byte, 32*1024)
	lastPersist := time.Now()

	for {
		select {
		case <-ctx.Done():
			if repo != nil {
				_ = repo.UpdatePartsProgress(part.ID, part.CurrentByte)
			}
			return ctx.Err()
		default:
		}

		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			offset := part.StartByte + part.CurrentByte
			if _, err := file.WriteAt(buf[:n], offset); err != nil {
				return fmt.Errorf("write error at offset %d: %w", offset, err)
			}

			part.CurrentByte += int64(n)
			totalDownloaded.Add(int64(n))

			// Periodically save part progress to SQLite (every 500ms)
			if repo != nil && time.Since(lastPersist) > 500*time.Millisecond {
				_ = repo.UpdatePartsProgress(part.ID, part.CurrentByte)
				lastPersist = time.Now()
			}
		}

		if readErr != nil {
			if readErr == io.EOF {
				if repo != nil {
					_ = repo.UpdatePartsProgress(part.ID, part.CurrentByte)
				}

				return nil
			}

			return readErr
		}
	}
}
