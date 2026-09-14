package downloader

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func FormatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTP"[exp])
}

func FormatTime(d time.Duration) string {
	totaltime := int64(d.Seconds())
	hour := totaltime / 3600
	remain := totaltime % 3600
	minute := remain / 60
	second := remain % 60

	return fmt.Sprintf("%02d:%02d:%02d", hour, minute, second)
}

func Truncate(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}

func resolveFilename(url string, opts DownloadOptions, resp *http.Response) string {
	var filename string
	if opts.Filename != "" {
		filename = opts.Filename
	} else if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		_, params, err := mime.ParseMediaType(cd)
		if err == nil && params["filename"] != "" {
			return params["filename"]
		}
	}

	if filename == "" {
		filename = path.Base(url)
	}

	return filepath.Base(filename)
}

func GetUniqueFilename(filename string) string {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return filename
	}

	ext := filepath.Ext(filename)
	name := strings.TrimSuffix(filename, ext)
	counter := 1

	for {
		newFilename := fmt.Sprintf("%s(%d)%s", name, counter, ext)
		if _, err := os.Stat(newFilename); os.IsNotExist(err) {
			return newFilename
		}
		counter++
	}
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

func ResolveDestination(filename string, categories []*Category, defaultDir string) (string, string) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))

	for _, cat := range categories {
		for _, catExt := range cat.Extension {
			if strings.EqualFold(ext, catExt) {
				_ = os.MkdirAll(cat.FolderPath, 0o755)
				return filepath.Join(cat.FolderPath, filename), cat.Name
			}
		}
	}

	// Fallback to default downloads directory
	_ = os.MkdirAll(defaultDir, 0o755)
	return filepath.Join(defaultDir, filename), "General"
}

func ParseRateLimit(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" || s == "0" || s == "UNLIMITED" {
		return 0, nil
	}

	multiplier := int64(1)
	if strings.HasSuffix(s, "M") || strings.HasSuffix(s, "KB") {
		multiplier = 1024
		s = strings.TrimSuffix(strings.TrimSuffix(s, "B"), "K")
	} else if strings.HasSuffix(s, "M") || strings.HasSuffix(s, "MB") {
		multiplier = 1024 * 1024
		s = strings.TrimSuffix(strings.TrimSuffix(s, "B"), "M")
	} else if strings.HasSuffix(s, "G") || strings.HasSuffix(s, "GB") {
		multiplier = 1024 * 1024 * 1024
		s = strings.TrimSuffix(strings.TrimSuffix(s, "B"), "G")
	}

	val, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid rate limit `%s`: %w", s, err)
	}

	return val * multiplier, nil 
}
