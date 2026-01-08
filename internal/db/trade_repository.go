package db

import (
	"fmt"
)

func (db *DB) SaveTrade(t *Trade) error {
	query := `INSERT INTO trades (id, trader_id, market_id, type, side, price, size, timestamp)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			  ON CONFLICT(id) DO UPDATE SET
			  trader_id=excluded.trader_id,
			  market_id=excluded.market_id,
			  type=excluded.type,
			  side=excluded.side,
			  price=excluded.price,
			  size=excluded.size,
			  timestamp=excluded.timestamp`
	
	_, err := db.conn.Exec(query, t.ID, t.TraderID, t.MarketID, t.Type, t.Side, t.Price, t.Size, t.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to save trade: %w", err)
	}
	return nil
}

func (db *DB) GetTradesByTrader(traderID string) ([]Trade, error) {
	query := `SELECT id, trader_id, market_id, type, side, price, size, timestamp FROM trades WHERE trader_id = ? ORDER BY timestamp DESC`
	rows, err := db.conn.Query(query, traderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trades by trader: %w", err)
	}
	defer rows.Close()

	var trades []Trade
	for rows.Next() {
		var t Trade
		if err := rows.Scan(&t.ID, &t.TraderID, &t.MarketID, &t.Type, &t.Side, &t.Price, &t.Size, &t.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan trade: %w", err)
		}
		trades = append(trades, t)
	}
	return trades, nil
}
