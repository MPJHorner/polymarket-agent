package cmd

import (

	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan Polymarket for high-performing traders",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("Scanning Polymarket...")
		// Implementation will go here
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
