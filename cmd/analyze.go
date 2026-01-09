package cmd

import (
	"context"
	"errors"
	"fmt"
	"polytracker/internal/claude"
	"polytracker/internal/db"
	"polytracker/internal/polymarket"

	"github.com/spf13/cobra"
)

var skipFetch bool

var analyzeCmd = &cobra.Command{
	Use:   "analyze [address]",
	Short: "Analyze a specific trader's strategy using Claude AI",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		address := args[0]

		database, err := db.NewDB(cfg.Database.Path)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		defer database.Close()

		// Fetch data unless --skip-fetch is set
		if !skipFetch {
			cmd.Printf("Fetching history for trader: %s\n", address)

			pmClient := polymarket.NewClient(polymarket.Config{
				APIKey:     cfg.Polymarket.APIKey,
				APISecret:  cfg.Polymarket.APISecret,
				Passphrase: cfg.Polymarket.Passphrase,
			})

			fetcher := polymarket.NewFetcher(pmClient, database)

			if err := fetcher.FetchTraderHistory(context.Background(), address); err != nil {
				return fmt.Errorf("fetch failed: %w", err)
			}
			cmd.Println("Data fetch complete.")
		}

		// Get trader data from database
		trader, err := database.GetTrader(address)
		if err != nil {
			return fmt.Errorf("failed to get trader: %w", err)
		}
		if trader == nil {
			return fmt.Errorf("trader not found: %s", address)
		}

		// Get trades
		trades, err := database.GetTradesByTrader(address)
		if err != nil {
			return fmt.Errorf("failed to get trades: %w", err)
		}

		// Get markets for trades
		markets := make(map[string]*db.Market)
		for _, trade := range trades {
			if _, exists := markets[trade.MarketID]; !exists {
				market, err := database.GetMarket(trade.MarketID)
				if err != nil {
					cmd.Printf("Warning: failed to get market %s: %v\n", trade.MarketID, err)
					continue
				}
				markets[trade.MarketID] = market
			}
		}

		// Check if Claude API key is configured
		if cfg.Claude.APIKey == "" {
			cmd.Println("\nClaude API key not configured. Skipping AI analysis.")
			cmd.Println("Set POLYTRACKER_CLAUDE_API_KEY or add claude.api_key to config.yaml")
			return nil
		}

		// Initialize Claude client
		claudeClient, err := claude.NewClient(claude.Config{
			APIKey:   cfg.Claude.APIKey,
			Endpoint: cfg.Claude.Endpoint,
		})
		if err != nil {
			return fmt.Errorf("failed to initialize Claude client: %w", err)
		}

		cmd.Printf("\nAnalyzing trader with Claude AI...\n")

		// Perform analysis
		result, err := claudeClient.AnalyzeTrader(context.Background(), claude.TraderData{
			Trader:  trader,
			Trades:  trades,
			Markets: markets,
		})

		if err != nil {
			if errors.Is(err, claude.ErrTokenLimit) {
				cmd.Println("Warning: Response was truncated due to token limit")
			} else {
				return fmt.Errorf("analysis failed: %w", err)
			}
		}

		// Display results
		cmd.Printf("\n%s\n", result.Thesis)
		cmd.Printf("\n---\n")
		cmd.Printf("Model: %s | Tokens: %d in / %d out\n", result.Model, result.InputTokens, result.OutputTokens)

		// Save analysis to database
		analysis := &db.Analysis{
			TraderID:  address,
			Thesis:    result.Thesis,
			CreatedAt: result.CreatedAt,
		}
		if err := database.SaveAnalysis(analysis); err != nil {
			cmd.Printf("Warning: failed to save analysis: %v\n", err)
		}

		return nil
	},
}

func init() {
	analyzeCmd.Flags().BoolVar(&skipFetch, "skip-fetch", false, "Skip fetching new data and use cached data only")
	rootCmd.AddCommand(analyzeCmd)
}

