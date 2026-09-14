package downloader

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"
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
	if strings.HasSuffix(s, "K") || strings.HasSuffix(s, "KB") {
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

func ParseScheduleTime(atStr, inStr string, now time.Time) (*time.Time, error) {
	if atStr != "" && inStr != "" {
		return nil, fmt.Errorf("cannot use both --at and --in flags simultaneously")
	}
	if atStr == "" && inStr == "" {
		return nil, nil
	}

	// 1. Relative duration: --in "30m", "2h"
	if inStr != "" {
		dur, err := time.ParseDuration(inStr)
		if err != nil {
			return nil, fmt.Errorf("invalid duration format '%s': %w", inStr, err)
		}
		t := now.Add(dur)
		return &t, nil
	}

	// 2. Absolute time: --at "15:04" or "2006-01-02 15:04"
	if t, err := time.ParseInLocation("2006-01-02 15:04", atStr, now.Location()); err == nil {
		return &t, nil
	}

	// Try time-only format: "15:04"
	if t, err := time.ParseInLocation("15:04", atStr, now.Location()); err == nil {
		target := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location())
		if target.Before(now) {
			target = target.Add(24 * time.Hour)
		}
		return &target, nil
	}

	return nil, fmt.Errorf("invalid time format '%s' (use '15:04' or '2006-01-02 15:04')", atStr)
}

func FallbackFilenameFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err == nil && parsed.Path != "" {
		base := filepath.Base(parsed.Path)
		if base != "" && base != "." && base != "/" {
			return filepath.Base(base)
		}
	}
	return "download"
}
