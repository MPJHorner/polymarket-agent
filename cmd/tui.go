package cmd

import (
	"log"
	"polytracker/internal/db"
	"polytracker/internal/ui"

	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the interactive terminal UI",
	Run: func(cmd *cobra.Command, args []string) {
		database, err := db.NewDB(cfg.Database.Path)
		if err != nil {
			log.Fatalf("Error opening database: %v", err)
		}
		defer database.Close()

		if err := ui.StartWithDB(cfg.UI.Theme, database); err != nil {
			log.Fatalf("Error starting TUI: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
