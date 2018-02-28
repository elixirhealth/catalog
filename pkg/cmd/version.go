package cmd

import (
	"os"

	"github.com/elxirhealth/catalog/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print the catalog version",
	Long:  "print the catalog version",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := os.Stdout.WriteString(version.Current.Version.String() + "\n")
		return err
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
