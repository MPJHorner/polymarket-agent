package polymarket

import (
	"context"
	"fmt"
	"log"
	"polytracker/internal/db"
	"time"
)

type Fetcher struct {
	client *Client
	db     *db.DB
}

func NewFetcher(client *Client, database *db.DB) *Fetcher {
	return &Fetcher{
		client: client,
		db:     database,
	}
}

// FetchTraderHistory performs a deep-dive fetch of all trades for a specific address
func (f *Fetcher) FetchTraderHistory(ctx context.Context, address string) error {
	log.Printf("Fetching history for trader: %s", address)

	// 1. Fetch account trades from Polymarket CLOB
	apiTrades, err := f.client.GetAccountTrades(ctx, address)
	if err != nil {
		return fmt.Errorf("failed to fetch account trades: %w", err)
	}

	for _, at := range apiTrades {
		// 2. Caching logic: check if we already have this trade
		// (SaveTrade uses ON CONFLICT DO UPDATE, but we can skip if we want to be more efficient)
		// For now, we'll just process it.

		// 3. Ensure market info is in DB
		if err := f.ensureMarket(ctx, at.MarketID); err != nil {
			log.Printf("Warning: failed to ensure market %s: %v", at.MarketID, err)
			continue
		}

		// 4. Map API trade to DB trade
		t := &db.Trade{
			ID:        at.ID,
			TraderID:  address,
			MarketID:  at.MarketID,
			Type:      at.Side, // BUY/SELL
			Price:     at.Price,
			Size:      at.Size,
			Timestamp: time.Unix(at.Timestamp, 0),
		}
		
		// Note: Side (YES/NO) is tricky without knowing which token was traded.
		// For now, we'll default to YES or try to infer if we had more info.
		t.Side = "YES" 

		if err := f.db.SaveTrade(t); err != nil {
			log.Printf("Error saving trade %s: %v", t.ID, err)
			continue
		}

		// 5. Fetch and store market snapshot for this trade time
		// Ideally we'd get a snapshot AT the trade time, but for now we'll just 
		// get the current market state as a snapshot if we don't have one recently.
		if err := f.ensureSnapshot(ctx, at.MarketID); err != nil {
			log.Printf("Warning: failed to ensure snapshot for market %s: %v", at.MarketID, err)
		}
	}

	return nil
}

func (f *Fetcher) ensureMarket(ctx context.Context, marketID string) error {
	m, err := f.db.GetMarket(marketID)
	if err != nil {
		return err
	}
	if m != nil {
		return nil // Already in DB
	}

	// Fetch from Gamma API
	apiMarket, err := f.client.GetMarket(ctx, marketID)
	if err != nil {
		return err
	}

	dbMarket := &db.Market{
		ID:       apiMarket.ID,
		Question: apiMarket.Question,
		Status:   "active", // Default
	}
	if apiMarket.Closed {
		dbMarket.Status = "closed"
	}

	return f.db.SaveMarket(dbMarket)
}

func (f *Fetcher) ensureSnapshot(ctx context.Context, marketID string) error {
	// Check if we have a recent snapshot (e.g., within the last hour)
	latest, err := f.db.GetLatestMarketSnapshot(marketID)
	if err != nil {
		return err
	}

	if latest != nil && time.Since(latest.Timestamp) < time.Hour {
		return nil
	}

	// Fetch current state from Gamma API
	apiMarket, err := f.client.GetMarket(ctx, marketID)
	if err != nil {
		return err
	}

	snapshot := &db.MarketSnapshot{
		MarketID:  marketID,
		Timestamp: time.Now(),
	}

	for _, token := range apiMarket.Tokens {
		if token.Outcome == "Yes" {
			snapshot.YesPrice = token.Price
		} else if token.Outcome == "No" {
			snapshot.NoPrice = token.Price
		}
	}

	return f.db.SaveMarketSnapshot(snapshot)
}
