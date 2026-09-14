package downloader

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func Download(id string, url string, opts DownloadOptions, ctx context.Context) (totalSize int64, filename string, err error) {
	client := &http.Client{}

	minMultiPartSize := int64(5 * 1024 * 1024)
	if opts.Settings != nil && opts.Settings.MinMultiPartDownload > 0 {
		minMultiPartSize = opts.Settings.MinMultiPartDownload
	}

	info, err := probeServerSupport(ctx, client, url, opts)
	if err != nil {
		return 0, "", fmt.Errorf("probe failed: %w", err)
	}

	baseName := info.Filename

	if opts.Dir != "" {
		// 1. User specified an explicit directory or path with -o or -d
		_ = os.MkdirAll(opts.Dir, 0o755)
		info.Filename = filepath.Join(opts.Dir, baseName)

	} else if filepath.IsAbs(opts.Filename) || filepath.Dir(opts.Filename) != "." {
		// User gave a specific relative or absolute path via -o (e.g. -o ./builds/app.bin)
		targetDir := filepath.Dir(opts.Filename)
		_ = os.MkdirAll(targetDir, 0o755)
		info.Filename = opts.Filename

	} else if opts.Settings != nil && opts.Settings.DownloadDir != "" && !filepath.IsAbs(info.Filename) {
		// 2. Default automatic category routing
		var categories []*Category
		if opts.Repo != nil {
			categories, _ = opts.Repo.GetCategories()
		}
		destPath, _ := ResolveDestination(info.Filename, categories, opts.Settings.DownloadDir)
		info.Filename = destPath
	}

	if info.SupportMultiPart && info.TotalSize > minMultiPartSize {
		return multipartDownload(id, url, info, opts, ctx)
	}

	return singlepartDownload(id, url, info, opts, ctx)
}
