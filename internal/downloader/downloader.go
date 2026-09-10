package downloader

import (
	"context"
	"fmt"
	"net/http"
)

func Download(id string, url string, opts DownloadOptions, ctx context.Context) (totalSize int64, filename string, err error) {
	client := &http.Client{}
	
	info, err := probeServerSupport(ctx, client, url, opts)
	if err != nil {
		return 0, "", fmt.Errorf("probe failed: %w", err)
	}

	const minMultiPartSize int64 = 5 * 1024 * 1024
	if info.SupportMultiPart && info.TotalSize > minMultiPartSize {
		return multipartDownload(id, url, info, opts, ctx)
	}

	return singlepartDownload(id, url, info, opts, ctx)
}
