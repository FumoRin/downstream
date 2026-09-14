package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/fumorin/gdl-manager/internal/downloader"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "View and modify gdl configuration",
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all of the current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := os.UserConfigDir()
		repo, err := downloader.NewSQLiteRepository(filepath.Join(dir, "gdl", "gdl.db"))
		if err != nil {
			return err
		}

		settings, err := repo.GetAllSettings()
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		_, _ = fmt.Fprintf(w, "KEY\tVALUE\n")
		_, _ = fmt.Fprintf(w, "download_dir\t%s\n", settings.DownloadDir)
		_, _ = fmt.Fprintf(w, "max_concurrent_download\t%d\n", settings.MaxConcurrencyDownload)
		_, _ = fmt.Fprintf(w, "per_part_download\t%d\n", settings.PartsPerDownload)
		_, _ = fmt.Fprintf(w, "min_multi_part_size\t%s\n", downloader.FormatBytes(settings.MinMultiPartDownload))
		return w.Flush()
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		val := args[1]

		dir, _ := os.UserConfigDir()
		repo, err := downloader.NewSQLiteRepository(filepath.Join(dir, "gdl", "gdl.db"))
		if err != nil {
			return err
		}

		if err := repo.SetSettings(key, val); err != nil {
			return fmt.Errorf("failed to save setting: %w", err)
		}

		fmt.Printf("Successfully set %s = %s\n", key, val)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configCmd)
}
