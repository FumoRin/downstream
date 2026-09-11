package downloader

import (
	"context"
	"fmt"
	"net/http"
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

	if info.SupportMultiPart && info.TotalSize > minMultiPartSize {
		return multipartDownload(id, url, info, opts, ctx)
	}

	return singlepartDownload(id, url, info, opts, ctx)
}
