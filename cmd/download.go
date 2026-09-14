package cmd

import (
	"fmt"
	"time"

	"github.com/fumorin/gdl-manager/internal/downloader"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"golang.org/x/time/rate"
)

var (
	filename     string
	rateLimitStr string
	atStr        string
	inStr        string
)

var downloadCmd = &cobra.Command{
	Use:   "download [url]",
	Short: "Download a file from the given url",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		m := app.Manager

		schedTime, err := downloader.ParseScheduleTime(atStr, inStr, time.Now())
		if err != nil {
			return err
		}

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

		if schedTime != nil {
			for _, url := range args {
				id := uuid.NewString()
				info, err := m.ScheduleDownload(id, url, filename, cmd.Context(), *schedTime, limiter)
				if err != nil {
					fmt.Printf("(offline fallback) Scheduled download for %s at %s \n", url, schedTime.Format("2006-01-02 15:04:05"))
				} else {
					// Successfully probed: display exact filename and human-readable size!
					fmt.Printf("Scheduled %s (%s) at %s\n",
						info.Filename,
						downloader.FormatBytes(info.TotalSize),
						schedTime.Format("2006-01-02 15:04:05"),
					)
				}
			}
			return nil
		}

		for _, url := range args {
			ID := uuid.NewString()
			m.StartDownload(ID, url, filename, limiter)
		}
		defer m.Close()

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
	downloadCmd.Flags().StringVar(&atStr, "at", "", "Schedule download at time (e.g. '23:00' or '2026-09-15 02:00')")
	downloadCmd.Flags().StringVar(&inStr, "in", "", "Schedule download after duration (e.g. '30m', '2h')")

	rootCmd.AddCommand(downloadCmd)
}
