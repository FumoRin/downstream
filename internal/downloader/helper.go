package downloader

import (
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	reFilenameStar = regexp.MustCompile(`(?i)filename\*=(?:["']?)([^';\s]+)'([^';]*)'([^;\s"]+)(?:["']?)`)
	reFilename     = regexp.MustCompile(`(?i)\bfilename\s*=\s*(?:"([^"]+)"|'([^']+)'|([^;\s]+))`)
)

var cp1252ToByte = map[rune]byte{
	0x20AC: 0x80, 0x201A: 0x82, 0x0192: 0x83, 0x201E: 0x84, 0x2026: 0x85,
	0x2020: 0x86, 0x2021: 0x87, 0x02C6: 0x88, 0x2030: 0x89, 0x0160: 0x8A,
	0x2039: 0x8B, 0x0152: 0x8C, 0x017D: 0x8E, 0x2018: 0x91, 0x2019: 0x92,
	0x201C: 0x93, 0x201D: 0x94, 0x2022: 0x95, 0x2013: 0x96, 0x2014: 0x97,
	0x02DC: 0x98, 0x2122: 0x99, 0x0161: 0x9A, 0x203A: 0x9B, 0x0153: 0x9C,
	0x017E: 0x9E, 0x0178: 0x9F,
}

func cleanFilename(name string) string {
	name = fixMojibake(name)
	name = strings.TrimSpace(name)
	name = filepath.Base(filepath.Clean(name))
	name = strings.Trim(name, `"' `)
	if name == "." || name == "/" || name == "\\" {
		return ""
	}
	return name
}

func fixMojibake(s string) string {
	if s == "" {
		return ""
	}
	hasExtended := false
	for _, r := range s {
		if r > 0x7F {
			hasExtended = true
			break
		}
	}
	if !hasExtended {
		return s
	}

	raw := make([]byte, 0, len(s))
	for _, r := range s {
		if r <= 0xFF {
			raw = append(raw, byte(r))
		} else if b, ok := cp1252ToByte[r]; ok {
			raw = append(raw, b)
		} else {
			return s
		}
	}

	if utf8.Valid(raw) {
		decoded := string(raw)
		if decoded != s {
			return decoded
		}
	}

	return s
}

func decodeFilename(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "=?") && strings.Contains(s, "?=") {
		dec := new(mime.WordDecoder)
		if decoded, err := dec.DecodeHeader(s); err == nil && decoded != "" {
			return cleanFilename(decoded)
		}
	}

	if strings.Contains(s, "%") {
		if unescaped, err := url.PathUnescape(s); err == nil && unescaped != "" {
			return cleanFilename(unescaped)
		}
		if unescaped, err := url.QueryUnescape(s); err == nil && unescaped != "" {
			return cleanFilename(unescaped)
		}
	}
	return cleanFilename(s)
}

func parseContentDisposition(header string) string {
	if header == "" {
		return ""
	}

	// 1. Highest priority: RFC 5987 / RFC 6266 filename* parameter
	if match := reFilenameStar.FindStringSubmatch(header); len(match) == 4 {
		charset := strings.ToLower(match[1])
		encoded := match[3]
		if charset == "utf-8" || charset == "" {
			if unescaped, err := url.PathUnescape(encoded); err == nil && unescaped != "" {
				if cleaned := cleanFilename(unescaped); cleaned != "" {
					return cleaned
				}
			}
			if unescaped, err := url.QueryUnescape(encoded); err == nil && unescaped != "" {
				if cleaned := cleanFilename(unescaped); cleaned != "" {
					return cleaned
				}
			}
		}
	}

	// 2. Try mime.ParseMediaType in case it parses standard parameters
	_, params, err := mime.ParseMediaType(header)
	if err == nil {
		if fnStar, ok := params["filename*"]; ok && fnStar != "" {
			parts := strings.SplitN(fnStar, "'", 3)
			if len(parts) == 3 && (strings.EqualFold(parts[0], "utf-8") || parts[0] == "") {
				if unescaped, err := url.PathUnescape(parts[2]); err == nil && unescaped != "" {
					if cleaned := cleanFilename(unescaped); cleaned != "" {
						return cleaned
					}
				}
			}
		}
		if fn, ok := params["filename"]; ok && fn != "" {
			if decoded := decodeFilename(fn); decoded != "" {
				return decoded
			}
		}
	}

	// 3. Fallback regex for filename="..." if mime.ParseMediaType failed on raw UTF-8
	if match := reFilename.FindStringSubmatch(header); len(match) > 0 {
		for i := 1; i < len(match); i++ {
			if match[i] != "" {
				if decoded := decodeFilename(match[i]); decoded != "" {
					return decoded
				}
			}
		}
	}

	return ""
}

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

func resolveFilename(urlStr string, opts DownloadOptions, resp *http.Response) string {
	if opts.Filename != "" {
		return cleanFilename(opts.Filename)
	}

	if resp != nil {
		if cd := resp.Header.Get("Content-Disposition"); cd != "" {
			if fn := parseContentDisposition(cd); fn != "" {
				return fn
			}
		}
	}

	return FallbackFilenameFromURL(urlStr)
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
	if err == nil {
		p := parsed.Path
		if p != "" && p != "/" {
			base := path.Base(p)
			if unescaped, err := url.PathUnescape(base); err == nil && unescaped != "" && unescaped != "." && unescaped != "/" {
				return cleanFilename(unescaped)
			}
			if unescaped, err := url.QueryUnescape(base); err == nil && unescaped != "" && unescaped != "." && unescaped != "/" {
				return cleanFilename(unescaped)
			}
			if base != "" && base != "." && base != "/" {
				return cleanFilename(base)
			}
		}
	}
	return "download"
}
