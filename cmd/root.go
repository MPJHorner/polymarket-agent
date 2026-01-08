package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"polytracker/internal/config"
)

var (
	cfgFile        string
	minTrades      int
	outputFile     string
	theme          string
	claudeEndpoint string
	cfg            *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "polytracker",
	Short: "Polytracker is a tool to track and analyze Polymarket traders",
	Long:  `A terminal application in Go to identify high-performing Polymarket traders, analyze their strategies using Claude AI, and track their performance.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.LoadConfig(cfgFile)
		if err != nil {
			return err
		}

		// Override config with flags if provided
		if theme != "" {
			cfg.UI.Theme = theme
		}
		if claudeEndpoint != "" {
			cfg.Claude.Endpoint = claudeEndpoint
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.polytracker/config.yaml)")
	rootCmd.PersistentFlags().IntVar(&minTrades, "min-trades", 0, "Minimum number of trades for scanner")
	rootCmd.PersistentFlags().StringVar(&outputFile, "output", "", "Output file path")
	rootCmd.PersistentFlags().StringVar(&theme, "theme", "", "UI theme (dracula, nord, etc.)")
	rootCmd.PersistentFlags().StringVar(&claudeEndpoint, "claude-endpoint", "", "Claude AI API endpoint")
}
