package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	RunE: func(cmd *cobra.Command, _ []string) error {
		fmt.Println("kuberadar version", version)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
