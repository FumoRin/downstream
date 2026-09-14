package downloader

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func supportsMultiPart(resp *http.Response) bool {
	if resp.ContentLength <= 0 {
		return false
	}

	if strings.EqualFold(resp.Header.Get("Accept-Ranges"), "bytes") {
		return true
	}

	if resp.Header.Get("Content-Range") != "" {
		return true
	}

	return false
}

func probeServerSupport(ctx context.Context, client *http.Client, url string, opts DownloadOptions) (*TargetInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err == nil {
		defer func() {
			_ = resp.Body.Close()
		}()
		if resp.StatusCode == http.StatusOK {
			return &TargetInfo{
				Filename:         resolveFilename(url, opts, resp),
				TotalSize:        resp.ContentLength,
				SupportMultiPart: supportsMultiPart(resp),
			}, nil
		}
	}

	getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	getReq.Header.Set("Range", "bytes=0-0")

	getResp, err := client.Do(getReq)
	if err != nil {
		return nil, err
	}

	defer func() {
		_, _ = io.Copy(io.Discard, getResp.Body)
		_ = getResp.Body.Close()
	}()

	filename := resolveFilename(url, opts, getResp)
	if getResp.StatusCode == http.StatusPartialContent {
		totalSize := parseTotalSize(getResp.Header.Get("Content-Range"))
		return &TargetInfo{
			Filename:         filename,
			TotalSize:        totalSize,
			SupportMultiPart: supportsMultiPart(getResp),
		}, nil
	}

	return &TargetInfo{
		Filename:         filename,
		TotalSize:        getResp.ContentLength,
		SupportMultiPart: false,
	}, nil
}

func parseTotalSize(contentRange string) int64 {
	parts := strings.Split(contentRange, "/")
	if len(parts) < 2 {
		return 0
	}

	size, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64)
	if err != nil {
		return 0
	}

	return size
}
