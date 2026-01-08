package db

import (
	"database/sql"
	"fmt"
)

func (db *DB) SaveMarket(m *Market) error {
	query := `INSERT INTO markets (id, question, description, category, ends_at, status)
			  VALUES (?, ?, ?, ?, ?, ?)
			  ON CONFLICT(id) DO UPDATE SET
			  question=excluded.question,
			  description=excluded.description,
			  category=excluded.category,
			  ends_at=excluded.ends_at,
			  status=excluded.status`
	
	_, err := db.conn.Exec(query, m.ID, m.Question, m.Description, m.Category, m.EndsAt, m.Status)
	if err != nil {
		return fmt.Errorf("failed to save market: %w", err)
	}
	return nil
}

func (db *DB) GetMarket(id string) (*Market, error) {
	query := `SELECT id, question, description, category, ends_at, status FROM markets WHERE id = ?`
	row := db.conn.QueryRow(query, id)

	var m Market
	err := row.Scan(&m.ID, &m.Question, &m.Description, &m.Category, &m.EndsAt, &m.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get market: %w", err)
	}
	return &m, nil
}

func (db *DB) SaveMarketSnapshot(s *MarketSnapshot) error {
	query := `INSERT INTO market_snapshots (market_id, yes_price, no_price, timestamp)
			  VALUES (?, ?, ?, ?)`
	
	_, err := db.conn.Exec(query, s.MarketID, s.YesPrice, s.NoPrice, s.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to save market snapshot: %w", err)
	}
	return nil
}

func (db *DB) GetLatestMarketSnapshot(marketID string) (*MarketSnapshot, error) {
	query := `SELECT id, market_id, yes_price, no_price, timestamp FROM market_snapshots 
			  WHERE market_id = ? ORDER BY timestamp DESC LIMIT 1`
	row := db.conn.QueryRow(query, marketID)

	var s MarketSnapshot
	err := row.Scan(&s.ID, &s.MarketID, &s.YesPrice, &s.NoPrice, &s.Timestamp)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest market snapshot: %w", err)
	}
	return &s, nil
}
