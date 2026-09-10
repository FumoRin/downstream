package downloader

import (
	"net/http"
	"testing"
	"time"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.00 KB"},
		{1536, "1.50 KB"},
		{1048576, "1.00 MB"},
		{1073741824, "1.00 GB"},
	}

	for _, tc := range tests {
		got := FormatBytes(tc.input)
		if got != tc.expected {
			t.Errorf("FormatBytes(%d) = %s, expected %s", tc.input, got, tc.expected)
		}
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		input    time.Duration
		expected string
	}{
		{0, "00:00:00"},
		{45 * time.Second, "00:00:45"},
		{2*time.Minute + 15*time.Second, "00:02:15"},
		{1*time.Hour + 23*time.Minute + 45*time.Second, "01:23:45"},
	}

	for _, tc := range tests {
		got := FormatTime(tc.input)
		if got != tc.expected {
			t.Errorf("FormatTime(%v) = %s, expected %s", tc.input, got, tc.expected)
		}
	}
}

func TestTruncate(t *testing.T) {
	if got := Truncate("short.txt", 20); got != "short.txt" {
		t.Errorf("expected 'short.txt', got '%s'", got)
	}

	longStr := "a_very_long_filename_that_should_be_truncated.zip"
	truncated := Truncate(longStr, 15)
	if len(truncated) != 15 || truncated[len(truncated)-3:] != "..." {
		t.Errorf("expected truncated string with '...', got '%s'", truncated)
	}
}

func TestResolveFilename(t *testing.T) {
	// Case 1: Custom option provided by user
	opts := DownloadOptions{Filename: "custom_name.tar"}
	resp := &http.Response{Header: make(http.Header)}
	if name := resolveFilename("http://example.com/original.tar", opts, resp); name != "custom_name.tar" {
		t.Errorf("expected 'custom_name.tar', got '%s'", name)
	}

	// Case 2: Content-Disposition header
	resp.Header.Set("Content-Disposition", `attachment; filename="server_suggested.pdf"`)
	if name := resolveFilename("http://example.com/download", DownloadOptions{}, resp); name != "server_suggested.pdf" {
		t.Errorf("expected 'server_suggested.pdf', got '%s'", name)
	}

	// Case 3: URL path fallback
	resp.Header.Del("Content-Disposition")
	if name := resolveFilename("http://example.com/images/wallpaper.png", DownloadOptions{}, resp); name != "wallpaper.png" {
		t.Errorf("expected 'wallpaper.png', got '%s'", name)
	}
}

func TestParseTotalSize(t *testing.T) {
	if size := parseTotalSize("bytes 0-0/1048576"); size != 1048576 {
		t.Errorf("expected 1048576, got %d", size)
	}
	if size := parseTotalSize("bytes 0-0/*"); size != 0 {
		t.Errorf("expected 0 for wildcard size, got %d", size)
	}
	if size := parseTotalSize(""); size != 0 {
		t.Errorf("expected 0 for empty string, got %d", size)
	}
}
