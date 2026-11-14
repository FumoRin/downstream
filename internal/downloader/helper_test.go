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

func TestParseRateLimit(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    int64
		expectError bool
	}{
		{"Empty string", "", 0, false},
		{"Zero", "0", 0, false},
		{"Unlimited lowercase", "unlimited", 0, false},
		{"Unlimited uppercase", "UNLIMITED", 0, false},
		{"Raw bytes", "2048", 2048, false},
		{"Kilobytes single letter", "500K", 500 * 1024, false},
		{"Kilobytes two letters", "500KB", 500 * 1024, false},
		{"Kilobytes lowercase", "500kb", 500 * 1024, false},
		{"Megabytes single letter", "2M", 2 * 1024 * 1024, false},
		{"Megabytes two letters", "2MB", 2 * 1024 * 1024, false},
		{"Megabytes lowercase", "2mb", 2 * 1024 * 1024, false},
		{"Gigabytes single letter", "1G", 1 * 1024 * 1024 * 1024, false},
		{"Gigabytes two letters", "1GB", 1 * 1024 * 1024 * 1024, false},
		{"Whitespace around", "  500K  ", 500 * 1024, false},
		{"Invalid text", "invalid", 0, true},
		{"Invalid unit format", "500MB/s", 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseRateLimit(tc.input)
			if tc.expectError {
				if err == nil {
					t.Errorf("ParseRateLimit(%q) expected error, got nil", tc.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseRateLimit(%q) unexpected error: %v", tc.input, err)
			}
			if got != tc.expected {
				t.Errorf("ParseRateLimit(%q) = %d, expected %d", tc.input, got, tc.expected)
			}
		})
	}
}

func TestResolveFilename_NonLatin(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		cdHeader string
		expected string
	}{
		{
			name:     "RFC 5987 filename* with UTF-8 and no lang",
			url:      "http://example.com/download",
			cdHeader: `attachment; filename*=UTF-8''%E6%97%A5%E6%9C%AC%E8%AA%9E.pdf`,
			expected: "日本語.pdf",
		},
		{
			name:     "RFC 5987 filename* with language tag",
			url:      "http://example.com/download",
			cdHeader: `attachment; filename*=UTF-8'ja'%E6%97%A5%E6%9C%AC%E8%AA%9E.pdf`,
			expected: "日本語.pdf",
		},
		{
			name:     "RFC 5987 filename* with fallback filename",
			url:      "http://example.com/download",
			cdHeader: `attachment; filename="fallback.pdf"; filename*=UTF-8''%E6%97%A5%E6%9C%AC%E8%AA%9E.pdf`,
			expected: "日本語.pdf",
		},
		{
			name:     "Percent-encoded standard filename",
			url:      "http://example.com/download",
			cdHeader: `attachment; filename="%E6%97%A5%E6%9C%AC%E8%AA%9E.pdf"`,
			expected: "日本語.pdf",
		},
		{
			name:     "Raw UTF-8 in quotes",
			url:      "http://example.com/download",
			cdHeader: `attachment; filename="日本語.pdf"`,
			expected: "日本語.pdf",
		},
		{
			name:     "RFC 2047 MIME encoded-word",
			url:      "http://example.com/download",
			cdHeader: `attachment; filename="=?UTF-8?B?5pel5pys6KqeLnBkZg==?="`,
			expected: "日本語.pdf",
		},
		{
			name:     "URL path fallback with percent-encoding",
			url:      "http://example.com/files/%E6%97%A5%E6%9C%AC%E8%AA%9E.pdf",
			cdHeader: "",
			expected: "日本語.pdf",
		},
		{
			name:     "URL path fallback with query parameters",
			url:      "http://example.com/files/%E6%97%A5%E6%9C%AC%E8%AA%9E.pdf?token=abc123xyz",
			cdHeader: "",
			expected: "日本語.pdf",
		},
		{
			name:     "Windows-1252 mojibake (Google Drive Japanese raw header)",
			url:      "http://example.com/download",
			cdHeader: `attachment; filename="ã‚·ãƒ£ãƒ³ãƒ†ã‚£(r-906 Remix).mp3"`,
			expected: "シャンティ(r-906 Remix).mp3",
		},
		{
			name:     "Raw ISO-8859-1 byte sequence (Google Drive C1 controls)",
			url:      "http://example.com/download",
			cdHeader: "attachment; filename=\"\xe3\x82\xb7\xe3\x83\xa3\xe3\x83\xb3\xe3\x83\x86\xe3\x82\xa3(r-906 Remix).mp3\"",
			expected: "シャンティ(r-906 Remix).mp3",
		},
		{
			name:     "Cyrillic Russian raw ISO-8859-1 header",
			url:      "http://example.com/download",
			cdHeader: "attachment; filename=\"\xd0\x9f\xd1\x80\xd0\xb8\xd0\xb2\xd0\xb5\xd1\x82_\xd0\xbc\xd0\xb8\xd1\x80.zip\"",
			expected: "Привет_мир.zip",
		},
		{
			name:     "Simplified Chinese raw ISO-8859-1 header",
			url:      "http://example.com/download",
			cdHeader: "attachment; filename=\"\xe4\xb8\xad\xe6\x96\x87\xe7\xae\x80\xe4\xbd\x93\xe6\xb5\x8b\xe8\xaf\x95.tar.gz\"",
			expected: "中文简体测试.tar.gz",
		},
		{
			name:     "Traditional Chinese raw ISO-8859-1 header",
			url:      "http://example.com/download",
			cdHeader: "attachment; filename=\"\xe7\xb9\x81\xe9\xab\x94\xe4\xb8\xad\xe6\x96\x87\xe6\xb8\xac\xe8\xa9\xa6.pdf\"",
			expected: "繁體中文測試.pdf",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := &http.Response{Header: make(http.Header)}
			if tc.cdHeader != "" {
				resp.Header.Set("Content-Disposition", tc.cdHeader)
			}
			got := resolveFilename(tc.url, DownloadOptions{}, resp)
			if got != tc.expected {
				t.Errorf("expected '%s', got '%s'", tc.expected, got)
			}
		})
	}
}
