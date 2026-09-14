package cmd

import (
	"fmt"

	"github.com/fumorin/gdl-manager/internal/downloader"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"golang.org/x/time/rate"
)

var filename string
var rateLimitStr string

var downloadCmd = &cobra.Command{
	Use:   "download [url]",
	Short: "Download a file from the given url",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		m := app.Manager

		if filename != "" && len(args) > 1 {
			return fmt.Errorf("error: Flag output is only available for single download")
		}

		bytePerSec, err := downloader.ParseRateLimit(rateLimitStr)
		if err != nil {
			return fmt.Errorf("invalid rate limit: %w", err)
		}

		var limiter *rate.Limiter
		if bytePerSec > 0 {
			limiter = rate.NewLimiter(rate.Limit(bytePerSec), 32*1024)
		}

		for _, url := range args {
			ID := uuid.NewString()
			m.StartDownload(ID, url, filename, limiter)
		}
		m.Close()

		return RunDownloadSession(m)
	},
}

func init() {
	downloadCmd.Flags().StringVarP(
		&filename,
		"output",
		"o",
		"",
		"The output filename provided by user (Only for a single download)",
	)
	downloadCmd.Flags().StringVarP(&rateLimitStr, "limit", "l", "", "Limit download speed (e.g. 500K, 2MB, etc)")
	rootCmd.AddCommand(downloadCmd)
}
