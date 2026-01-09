package cmd

import (
	"log"

	"polytracker/internal/claude"
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

		// Try to create a Claude client if API key is configured
		var claudeClient *claude.Client
		if cfg.Claude.APIKey != "" {
			claudeClient, err = claude.NewClient(claude.Config{
				APIKey:   cfg.Claude.APIKey,
				Endpoint: cfg.Claude.Endpoint,
			})
			if err != nil {
				// Log warning but don't fail - analysis just won't be available
				log.Printf("Warning: Could not initialize Claude client: %v", err)
			}
		}

		if claudeClient != nil {
			if err := ui.StartWithClaudeClient(cfg.UI.Theme, database, claudeClient); err != nil {
				log.Fatalf("Error starting TUI: %v", err)
			}
		} else {
			if err := ui.StartWithDB(cfg.UI.Theme, database); err != nil {
				log.Fatalf("Error starting TUI: %v", err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
