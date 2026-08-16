package downloader

import (
	"context"
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
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

	acceptedRanges := strings.ToLower(resp.Header.Get("Accepted-Ranges"))
	if acceptedRanges == "bytes" {
		return true
	}

	if resp.Header.Get("Content-Range") != ""{
		return true
	}

	return false
}

func probeServerSupport (ctx context.Context, client *http.Client, url string) (totalSize int64, multiPartSupported bool, filename string, err error) {
	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return 0, false, "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, false, "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusOK {
		totalSize = resp.ContentLength
		multiPartSupported = supportsMultiPart(resp)
		filename = resolveFilename(url, DownloadOptions{}, resp)
		return totalSize, multiPartSupported, filename, nil
	}
	
	getReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, false, "", err
	}

	getReq.Header.Set("Range", "bytes=0-0")

	getResp, err := client.Do(getReq)
	if err != nil {
		return 0, false, "", err
	}
	defer getResp.Body.Close()

	if getResp.StatusCode == http.StatusPartialContent {
		multiPartSupported = true
	}

	return totalSize, multiPartSupported, filename, nil
}
