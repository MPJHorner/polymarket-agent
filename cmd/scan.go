package cmd

import (
	"context"
	"log"
	"polytracker/internal/db"
	"polytracker/internal/polymarket"

	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan Polymarket for high-performing traders",
	Run: func(cmd *cobra.Command, args []string) {
		database, err := db.NewDB(cfg.Database.Path)
		if err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
		defer database.Close()

		client := polymarket.NewClient(polymarket.Config{
			APIKey:     cfg.Polymarket.APIKey,
			APISecret:  cfg.Polymarket.APISecret,
			Passphrase: cfg.Polymarket.Passphrase,
		})

		scanner := polymarket.NewScanner(client, database)
		
		cmd.Println("Scanning Polymarket for recent activity...")
		// Use a default limit or a flag if implemented
		limit := 10
		if err := scanner.ScanRecentActivity(context.Background(), limit); err != nil {
			log.Fatalf("Scan failed: %v", err)
		}
		cmd.Println("Scan complete.")
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
