package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func singlepartDownload(id string, url string, info *TargetInfo, opts DownloadOptions, ctx context.Context) (int64, string, error) {
	filename := info.Filename
	tmpFilename := filename + "." + id + ".tmp"
	client := &http.Client{}

	var currentSize int64
	if stat, err := os.Stat(tmpFilename); err == nil {
		currentSize = stat.Size()
	}

	for attempt := range defaultMaxRetries {
		if ctx.Err() != nil {
			return 0, "", ctx.Err()
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return 0, "", err
		}
		
		if currentSize > 0 {
			req.Header.Add("Range", fmt.Sprintf("bytes=%d-", currentSize))
		}

		resp, err := client.Do(req)
		if err != nil {
			if attempt == defaultMaxRetries {
				return 0, "", fmt.Errorf("exceeded max retries: %w", err)
			}
			_ = sleepWithContext(ctx, calculateBackOff(attempt, defaultBaseDelay, defaultMaxDelay))
			continue
		}

		var file *os.File
		var totalsize int64
		switch resp.StatusCode {
		case http.StatusPartialContent:
			totalsize = currentSize + resp.ContentLength
			file, err = os.OpenFile(tmpFilename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		case http.StatusOK:
			totalsize = resp.ContentLength
			file, err = os.Create(tmpFilename)
			currentSize = 0
		default:
			_ = resp.Body.Close()
			return 0, "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}

		if err != nil {
			_ = resp.Body.Close()
			return 0, "", err
		}

		pw := &ProgressWriter{
			Filename: filename,
			Total: totalsize,
			Current: currentSize,
			ByteAtStart: currentSize,
			Destination: file,
			StartTime: time.Now(),
			ProgressChan: opts.Progress,
		}

		buf := make([]byte, 32*1024)
		throttleBody := NewThrottledReader(ctx, resp.Body, opts.Limiter)
		_, copyErr := io.CopyBuffer(pw, throttleBody, buf)
		_ = resp.Body.Close()
		_ = file.Close()

		if copyErr == nil {
			finalFilename := GetUniqueFilename(filename)
			if err := os.Rename(tmpFilename, finalFilename); err != nil {
				return totalsize, filename, err
			}
			return totalsize, finalFilename, err
		}
		
		if ctx.Err() != nil {
			return 0, "", ctx.Err()
		}
		
		if stat, err := os.Stat(tmpFilename); err == nil {
			currentSize = stat.Size()
		}

		if attempt == defaultMaxRetries {
			return 0, "", fmt.Errorf("single-part failed after %d retries: %w", defaultMaxRetries, copyErr)
		}

		_ = sleepWithContext(ctx, calculateBackOff(attempt, defaultBaseDelay, defaultMaxDelay))
	}

	return 0, "", fmt.Errorf("download failed")
	}
