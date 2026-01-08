package cmd

import (
	"context"
	"fmt"
	"polytracker/internal/db"
	"polytracker/internal/polymarket"

	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [address]",
	Short: "Analyze a specific trader's strategy using Claude AI",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		address := args[0]
		cmd.Printf("Deep-dive fetching history for trader: %s\n", address)

		database, err := db.NewDB(cfg.Database.Path)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer database.Close()

		client := polymarket.NewClient(polymarket.Config{
			APIKey:     cfg.Polymarket.APIKey,
			APISecret:  cfg.Polymarket.APISecret,
			Passphrase: cfg.Polymarket.Passphrase,
		})

		fetcher := polymarket.NewFetcher(client, database)
		
		if err := fetcher.FetchTraderHistory(context.Background(), address); err != nil {
			return fmt.Errorf("fetch failed: %w", err)
		}

		cmd.Println("Data fetch complete. Ready for AI analysis.")
		// AI Analysis (ai-001) will be triggered here in the future
		return nil
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}

