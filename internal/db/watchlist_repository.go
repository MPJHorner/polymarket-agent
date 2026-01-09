package db

import (
	"database/sql"
	"fmt"
	"time"
)

func (db *DB) GetWatchlistItem(traderID string) (*WatchlistItem, error) {
	query := `SELECT trader_id, notes, created_at FROM watchlist WHERE trader_id = ?`
	row := db.conn.QueryRow(query, traderID)

	var item WatchlistItem
	err := row.Scan(&item.TraderID, &item.Notes, &item.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get watchlist item: %w", err)
	}
	return &item, nil
}

func (db *DB) AddToWatchlist(traderID string, notes string) error {
	query := `INSERT INTO watchlist (trader_id, notes, created_at)
			  VALUES (?, ?, ?)
			  ON CONFLICT(trader_id) DO UPDATE SET
			  notes=excluded.notes`

	_, err := db.conn.Exec(query, traderID, notes, time.Now())
	if err != nil {
		return fmt.Errorf("failed to add to watchlist: %w", err)
	}
	return nil
}

func (db *DB) RemoveFromWatchlist(traderID string) error {
	query := `DELETE FROM watchlist WHERE trader_id = ?`

	_, err := db.conn.Exec(query, traderID)
	if err != nil {
		return fmt.Errorf("failed to remove from watchlist: %w", err)
	}
	return nil
}

func (db *DB) ListWatchlist() ([]WatchlistItem, error) {
	query := `SELECT trader_id, notes, created_at FROM watchlist ORDER BY created_at DESC`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list watchlist: %w", err)
	}
	defer rows.Close()

	var items []WatchlistItem
	for rows.Next() {
		var item WatchlistItem
		if err := rows.Scan(&item.TraderID, &item.Notes, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan watchlist item: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}
