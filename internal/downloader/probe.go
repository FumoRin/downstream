package downloader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// applyHeaders injects custom headers (e.g. Cookie, Authorization, User-Agent) into an outgoing http.Request
func applyHeaders(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		req.Header.Set(k, v)
	}
}

// ProbeURLWithHeaders tests reachability using provided request headers (cookies, auth tokens, etc.)
func ProbeURLWithHeaders(ctx context.Context, rawURL string, headers map[string]string) (*TargetInfo, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	opts := DownloadOptions{
		URL:     rawURL,
		Headers: headers,
	}
	return probeServerSupport(ctx, client, rawURL, opts)
}

// ProbeURL tests reachability and retrieves remote file metadata (filename, size, multi-part support)
func ProbeURL(ctx context.Context, rawURL string) (*TargetInfo, error) {
	return ProbeURLWithHeaders(ctx, rawURL, nil)
}

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
	applyHeaders(req, opts.Headers)

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

	applyHeaders(getReq, opts.Headers)
	getReq.Header.Set("Range", "bytes=0-0")

	getResp, err := client.Do(getReq)
	if err != nil {
		return nil, err
	}

	defer func() {
		_, _ = io.Copy(io.Discard, getResp.Body)
		_ = getResp.Body.Close()
	}()

	if getResp.StatusCode >= 400 {
		return nil, fmt.Errorf("probe failed with status: %d %s", getResp.StatusCode, http.StatusText(getResp.StatusCode))
	}

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
