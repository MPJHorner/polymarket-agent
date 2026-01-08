package polymarket

import (
	"context"
	"fmt"
	"log"
	"polytracker/internal/db"
	"time"
)

type Scanner struct {
	client *Client
	db     *db.DB
}

func NewScanner(client *Client, database *db.DB) *Scanner {
	return &Scanner{
		client: client,
		db:     database,
	}
}

// ScanRecentActivity fetches recent markets and their trades to identify active traders
func (s *Scanner) ScanRecentActivity(ctx context.Context, marketLimit int) error {
	markets, err := s.client.ListMarkets(ctx, marketLimit)
	if err != nil {
		return fmt.Errorf("failed to list markets: %w", err)
	}

	traderStats := make(map[string]*db.Trader)

	for _, m := range markets {
		// Polymarket Gamma API returns conditionId which is often used in CLOB
		// But GetTrades expects marketID (which might be the same as conditionID or slug)
		// For CLOB API, we usually need the token ID or similar.
		// Let's use the ID for now, but we might need to adjust based on how CLOB API works.
		
		log.Printf("Scanning market: %s", m.Question)
		trades, err := s.client.GetTrades(ctx, m.ID)
		if err != nil {
			log.Printf("Warning: failed to get trades for market %s: %v", m.ID, err)
			continue
		}

		for _, t := range trades {
			s.processTrade(t, traderStats)
		}
	}

	// Save traders to database
	for _, trader := range traderStats {
		if err := s.db.SaveTrader(trader); err != nil {
			log.Printf("Error saving trader %s: %v", trader.Address, err)
		}
	}

	return nil
}

func (s *Scanner) processTrade(t Trade, stats map[string]*db.Trader) {
	addresses := []string{t.Maker, t.Taker}
	volume := t.Price * t.Size

	for _, addr := range addresses {
		if addr == "" {
			continue
		}
		
		trader, ok := stats[addr]
		if !ok {
			trader = &db.Trader{
				Address:     addr,
				LastScanned: time.Now(),
			}
			stats[addr] = trader
		}
		
		trader.Volume += volume
		// In a real scenario, we'd need more data to calculate Win Rate and P&L
		// For now, we're just aggregating volume from the scanned trades.
		trader.LastScanned = time.Now()
	}
}
