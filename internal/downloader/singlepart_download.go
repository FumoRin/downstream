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

	var currentSize int64
	if stat, err := os.Stat(tmpFilename); err == nil {
		currentSize = stat.Size()
	}

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, "", err
	}
	
	if currentSize > 0 {
		req.Header.Set("Range", fmt.Sprintf("Bytes=%d-", currentSize))
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer func()  {
		_ = resp.Body.Close()
	}()

	var file *os.File
	var totalSize int64

	switch resp.StatusCode {
	case http.StatusPartialContent:
		totalSize = currentSize + resp.ContentLength
		file, err = os.OpenFile(tmpFilename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)

	case http.StatusOK:
		totalSize = resp.ContentLength
		file, err = os.Create(tmpFilename)
	default:
		return 0, "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	if err != nil {
		return 0, "", err
	}
	
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close file: %w", closeErr)
		}
	}()

	pw := &ProgressWriter{
		Filename: filename,
		Total: totalSize,
		Current: currentSize,
		ByteAtStart: currentSize,
		Destination: file,
		StartTime: time.Now(),
		ProgressChan: opts.Progress,
	}

	buf := make([]byte, 32*1024)
	if _, err := io.CopyBuffer(pw, resp.Body, buf); err != nil {
		return 0, "", err
	}

	finalFilename := GetUniqueFilename(filename)
	if err := os.Rename(tmpFilename, finalFilename); err != nil {
		return totalSize, filename, err
	}

	return totalSize, finalFilename, nil
	}
