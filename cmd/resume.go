package cmd

import (
	"fmt"

	"github.com/fumorin/gdl-manager/internal/downloader"
	"github.com/spf13/cobra"
	"golang.org/x/time/rate"
)

var resumeRateLimitStr string

var resumeCmd = &cobra.Command{
	Use:   "resume [id]",
	Short: "Resume a paused download",
	RunE: func(cmd *cobra.Command, args []string) error {
		m := app.Manager

		var limiter *rate.Limiter

		if resumeRateLimitStr != "" {
			bytesPerSec, err := downloader.ParseRateLimit(resumeRateLimitStr)
			if err != nil {
				return fmt.Errorf("invalid rate limit: %w", err)
			}
			if bytesPerSec > 0 {
				limiter = rate.NewLimiter(rate.Limit(bytesPerSec), 32*1024)
			}
		}

		if len(args) > 0 {
			// If args provided
			targetID := args[0]

			state, err := m.GetDownload(targetID)
			if err != nil {
				return fmt.Errorf("failed to query download: %v", err)
			}
			if state == nil {
				return fmt.Errorf("download with provided ID not found: %s", targetID)
			}

			if state.Status != downloader.StateCompleted &&
				state.Status != downloader.StateDownloading &&
				state.Status != downloader.StateQueue {
				fmt.Printf("Resuming specific download: %s\n", state.Filename)
				m.StartDownload(targetID, state.URL, state.Filename, limiter)
			} else {
				fmt.Printf("Download %s is already completed\n", state.Filename)
			}
		} else {
			// If no Args provided
			states, err := m.GetIncompleteDownload()
			if err != nil {
				return fmt.Errorf("failed to query incomplete downloads: %v", err)
			}

			if len(states) == 0 {
				fmt.Println("No paused or incomplete downloads to resume.")
				return nil
			}

			for _, state := range states {
				if state.Status == downloader.StateScheduled {
					continue // skip the scheduled one, let it trigger by itself
				}
				m.StartDownload(state.ID, state.URL, state.Filename, limiter)
			}
		}

		m.Close()
		return RunDownloadSession(m)
	},
}

func init() {
	resumeCmd.Flags().StringVarP(&resumeRateLimitStr, "limit", "l", "", "Limit download speed (e.g. 500K, 2MB, etc)")
	rootCmd.AddCommand(resumeCmd)
}
