package cmd

import (
	"fmt"

	"polytracker/internal/db"
	"polytracker/internal/export"

	"github.com/spf13/cobra"
)

var (
	exportType     string
	traderAddress  string
	exportFilename string
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export trader data and analyses",
	Long: `Export trader data to various formats.

Available export types:
  leaderboard  Export the trader leaderboard to CSV
  thesis       Export a trader's analysis thesis to Markdown

Examples:
  polytracker export --type leaderboard
  polytracker export --type thesis --trader 0x1234...
  polytracker export --type leaderboard --output my_export.csv`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize database
		database, err := db.NewDB(cfg.Database.Path)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer database.Close()

		// Initialize exporter
		exporter := export.NewExporter("exports")

		switch exportType {
		case "leaderboard", "csv":
			return exportLeaderboard(cmd, database, exporter)
		case "thesis", "analysis", "md":
			return exportThesis(cmd, database, exporter)
		default:
			return fmt.Errorf("unknown export type: %s (use 'leaderboard' or 'thesis')", exportType)
		}
	},
}

func exportLeaderboard(cmd *cobra.Command, database *db.DB, exporter *export.Exporter) error {
	cmd.Println("Exporting leaderboard to CSV...")

	traders, err := database.ListTradersWithOptions(db.ListTradersOptions{
		SortBy: db.SortByProfitLoss,
		Order:  db.SortDesc,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch traders: %w", err)
	}

	if len(traders) == 0 {
		cmd.Println("No traders found in database. Run 'polytracker scan' first.")
		return nil
	}

	path, err := exporter.ExportLeaderboardCSV(traders, exportFilename)
	if err != nil {
		return fmt.Errorf("failed to export leaderboard: %w", err)
	}

	cmd.Printf("Exported %d traders to: %s\n", len(traders), path)
	return nil
}

func exportThesis(cmd *cobra.Command, database *db.DB, exporter *export.Exporter) error {
	if traderAddress == "" {
		return fmt.Errorf("trader address required for thesis export (use --trader flag)")
	}

	cmd.Printf("Exporting thesis for trader: %s...\n", traderAddress)

	path, err := exporter.ExportAnalysisFromDB(database, traderAddress, exportFilename)
	if err != nil {
		return fmt.Errorf("failed to export thesis: %w", err)
	}

	cmd.Printf("Exported thesis to: %s\n", path)
	return nil
}

// ExportLeaderboardCSV is a convenience function for programmatic use
func ExportLeaderboardCSV(dbPath string, outputFile string) error {
	database, err := db.NewDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer database.Close()

	traders, err := database.ListTraders()
	if err != nil {
		return fmt.Errorf("failed to fetch traders: %w", err)
	}

	exporter := export.NewExporter("exports")
	_, err = exporter.ExportLeaderboardCSV(traders, outputFile)
	return err
}

// ExportThesisMarkdown is a convenience function for programmatic use
func ExportThesisMarkdown(dbPath string, traderAddress string, outputFile string) error {
	database, err := db.NewDB(dbPath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer database.Close()

	exporter := export.NewExporter("exports")
	_, err = exporter.ExportAnalysisFromDB(database, traderAddress, outputFile)
	return err
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVarP(&exportType, "type", "t", "leaderboard", "Export type (leaderboard, thesis)")
	exportCmd.Flags().StringVarP(&traderAddress, "trader", "a", "", "Trader address for thesis export")
	exportCmd.Flags().StringVarP(&exportFilename, "filename", "f", "", "Output filename (auto-generated if not specified)")

	// Also support the global --output flag
	exportCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if outputFile != "" && exportFilename == "" {
			exportFilename = outputFile
		}
		return nil
	}
}
