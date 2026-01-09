package export

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"polytracker/internal/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExporter(t *testing.T) {
	// Test with empty directory (should use default)
	e := NewExporter("")
	assert.Equal(t, "exports", e.exportDir)

	// Test with custom directory
	e = NewExporter("custom_exports")
	assert.Equal(t, "custom_exports", e.exportDir)
}

func TestExportLeaderboardCSV(t *testing.T) {
	// Create a temporary directory for exports
	tmpDir, err := os.MkdirTemp("", "export_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	exporter := NewExporter(tmpDir)

	// Create sample trader data
	traders := []db.Trader{
		{
			Address:     "0x1234567890abcdef1234567890abcdef12345678",
			Username:    "trader1",
			WinRate:     0.75,
			ProfitLoss:  1234.56,
			ROI:         0.25,
			Volume:      50000.00,
			LastScanned: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			Address:     "0xabcdef1234567890abcdef1234567890abcdef12",
			Username:    "trader2",
			WinRate:     0.60,
			ProfitLoss:  -500.00,
			ROI:         -0.05,
			Volume:      10000.00,
			LastScanned: time.Date(2024, 1, 16, 14, 45, 0, 0, time.UTC),
		},
	}

	// Test with auto-generated filename
	path, err := exporter.ExportLeaderboardCSV(traders, "")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(filepath.Base(path), "leaderboard_"))
	assert.True(t, strings.HasSuffix(path, ".csv"))

	// Verify file contents
	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	require.NoError(t, err)

	// Verify header
	assert.Equal(t, []string{"Rank", "Address", "Username", "Win Rate (%)", "P&L ($)", "ROI (%)", "Volume ($)", "Last Scanned"}, records[0])

	// Verify data rows
	assert.Len(t, records, 3) // header + 2 traders
	assert.Equal(t, "1", records[1][0])
	assert.Equal(t, "0x1234567890abcdef1234567890abcdef12345678", records[1][1])
	assert.Equal(t, "trader1", records[1][2])
	assert.Equal(t, "75.00", records[1][3])
	assert.Equal(t, "1234.56", records[1][4])
	assert.Equal(t, "25.00", records[1][5])
	assert.Equal(t, "50000.00", records[1][6])

	assert.Equal(t, "2", records[2][0])
	assert.Equal(t, "trader2", records[2][2])
	assert.Equal(t, "-500.00", records[2][4])
}

func TestExportLeaderboardCSVCustomFilename(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "export_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	exporter := NewExporter(tmpDir)

	traders := []db.Trader{
		{
			Address:     "0x1234",
			Username:    "test",
			WinRate:     0.5,
			ProfitLoss:  100,
			ROI:         0.1,
			Volume:      1000,
			LastScanned: time.Now(),
		},
	}

	// Test with custom filename without extension
	path, err := exporter.ExportLeaderboardCSV(traders, "my_export")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tmpDir, "my_export.csv"), path)

	// Test with custom filename with extension
	path, err = exporter.ExportLeaderboardCSV(traders, "another_export.csv")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tmpDir, "another_export.csv"), path)
}

func TestExportLeaderboardCSVEmpty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "export_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	exporter := NewExporter(tmpDir)

	// Export empty list
	path, err := exporter.ExportLeaderboardCSV([]db.Trader{}, "empty.csv")
	require.NoError(t, err)

	// Verify file only has header
	file, err := os.Open(path)
	require.NoError(t, err)
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	require.NoError(t, err)
	assert.Len(t, records, 1) // header only
}

func TestExportThesisMarkdown(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "export_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	exporter := NewExporter(tmpDir)

	trader := &db.Trader{
		Address:     "0x1234567890abcdef1234567890abcdef12345678",
		Username:    "successful_trader",
		WinRate:     0.80,
		ProfitLoss:  5000.00,
		ROI:         0.50,
		Volume:      100000.00,
		LastScanned: time.Now(),
	}

	thesis := `## Trading Strategy
This trader demonstrates a strong momentum-based strategy.

### Key Observations
1. High win rate indicates good market timing
2. Consistent profit-taking behavior
3. Focus on high-liquidity markets

### Risk Assessment
- Low drawdown periods
- Well-managed position sizes`

	// Test with auto-generated filename
	path, err := exporter.ExportThesisMarkdown(trader, thesis, "")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(filepath.Base(path), "thesis_0x12345678"))
	assert.True(t, strings.HasSuffix(path, ".md"))

	// Verify file contents
	content, err := os.ReadFile(path)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "# Trading Thesis: 0x1234567890abcdef1234567890abcdef12345678")
	assert.Contains(t, contentStr, "**Trader:** successful_trader")
	assert.Contains(t, contentStr, "| Win Rate | 80.00% |")
	assert.Contains(t, contentStr, "| P&L | $5000.00 |")
	assert.Contains(t, contentStr, "| ROI | 50.00% |")
	assert.Contains(t, contentStr, "| Volume | $100000.00 |")
	assert.Contains(t, contentStr, "## Analysis")
	assert.Contains(t, contentStr, "## Trading Strategy")
	assert.Contains(t, contentStr, "Generated by Polytracker")
}

func TestExportThesisMarkdownCustomFilename(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "export_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	exporter := NewExporter(tmpDir)

	trader := &db.Trader{
		Address:     "0x1234",
		WinRate:     0.5,
		ProfitLoss:  100,
		ROI:         0.1,
		Volume:      1000,
		LastScanned: time.Now(),
	}

	// Test with custom filename without extension
	path, err := exporter.ExportThesisMarkdown(trader, "Test thesis content", "my_thesis")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tmpDir, "my_thesis.md"), path)

	// Test with custom filename with extension
	path, err = exporter.ExportThesisMarkdown(trader, "Test thesis content", "another_thesis.md")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(tmpDir, "another_thesis.md"), path)
}

func TestExportThesisMarkdownEmptyThesis(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "export_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	exporter := NewExporter(tmpDir)

	trader := &db.Trader{
		Address:     "0x1234",
		WinRate:     0.5,
		ProfitLoss:  100,
		ROI:         0.1,
		Volume:      1000,
		LastScanned: time.Now(),
	}

	// Empty thesis should return error
	_, err = exporter.ExportThesisMarkdown(trader, "", "test.md")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no thesis content")
}

func TestExportAnalysisFromDB(t *testing.T) {
	// Create a temporary directory and database
	tmpDir, err := os.MkdirTemp("", "export_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	database, err := db.NewDB(dbPath)
	require.NoError(t, err)
	defer database.Close()

	// Add test data
	trader := &db.Trader{
		Address:     "0xtest_trader_address",
		Username:    "test_user",
		WinRate:     0.65,
		ProfitLoss:  2500.00,
		ROI:         0.30,
		Volume:      75000.00,
		LastScanned: time.Now(),
	}
	err = database.SaveTrader(trader)
	require.NoError(t, err)

	analysis := &db.Analysis{
		TraderID:  trader.Address,
		Thesis:    "This is a test thesis for the trader analysis.",
		CreatedAt: time.Now(),
	}
	err = database.SaveAnalysis(analysis)
	require.NoError(t, err)

	// Export the analysis
	exportDir := filepath.Join(tmpDir, "exports")
	exporter := NewExporter(exportDir)

	path, err := exporter.ExportAnalysisFromDB(database, trader.Address, "")
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(path)
	require.NoError(t, err)

	// Verify contents
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "0xtest_trader_address")
	assert.Contains(t, string(content), "test_user")
	assert.Contains(t, string(content), "This is a test thesis")
}

func TestExportAnalysisFromDBTraderNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "export_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	database, err := db.NewDB(dbPath)
	require.NoError(t, err)
	defer database.Close()

	exporter := NewExporter(tmpDir)

	_, err = exporter.ExportAnalysisFromDB(database, "0xnonexistent", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "trader not found")
}

func TestExportAnalysisFromDBNoAnalysis(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "export_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	database, err := db.NewDB(dbPath)
	require.NoError(t, err)
	defer database.Close()

	// Add trader without analysis
	trader := &db.Trader{
		Address:     "0xtrader_no_analysis",
		WinRate:     0.5,
		ProfitLoss:  100,
		ROI:         0.1,
		Volume:      1000,
		LastScanned: time.Now(),
	}
	err = database.SaveTrader(trader)
	require.NoError(t, err)

	exporter := NewExporter(tmpDir)

	_, err = exporter.ExportAnalysisFromDB(database, trader.Address, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no analysis found")
}

func TestEnsureExportDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "export_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	nestedDir := filepath.Join(tmpDir, "nested", "exports")
	exporter := NewExporter(nestedDir)

	// Directory should not exist yet
	_, err = os.Stat(nestedDir)
	assert.True(t, os.IsNotExist(err))

	// EnsureExportDir should create it
	err = exporter.EnsureExportDir()
	require.NoError(t, err)

	// Directory should now exist
	info, err := os.Stat(nestedDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Calling again should not error
	err = exporter.EnsureExportDir()
	require.NoError(t, err)
}
