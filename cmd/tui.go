package cmd

import (
	"log"
	"polytracker/internal/ui"

	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the interactive terminal UI",
	Run: func(cmd *cobra.Command, args []string) {
		if err := ui.Start(); err != nil {
			log.Fatalf("Error starting TUI: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
