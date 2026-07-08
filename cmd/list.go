package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all of the download job",
	RunE: func(cmd *cobra.Command, args []string) error {
		m := app.Manager
		states, err := m.GetAllDownload()
		if err != nil {
			return fmt.Errorf("failed to retrieve downloads: %v", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

		_, printErr := fmt.Fprintln(w, "ID\tFilename\tStatus")
		if printErr != nil {
			return printErr
		}

		for _, state := range states {
			_, printErr := fmt.Fprintf(w, "%s\t%s\t%s\n", state.ID, state.Filename, state.Status)
			if printErr != nil {
				return printErr
			}
		}
		return w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
