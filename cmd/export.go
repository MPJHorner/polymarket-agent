package cmd

import (

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export trader data and analyses",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("Exporting data...")
		// Implementation will go here
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
}
