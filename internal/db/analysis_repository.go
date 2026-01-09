package db

import (
	"database/sql"
	"fmt"
)

func (db *DB) SaveAnalysis(a *Analysis) error {
	query := `INSERT INTO analyses (trader_id, thesis, created_at)
			  VALUES (?, ?, ?)`

	result, err := db.conn.Exec(query, a.TraderID, a.Thesis, a.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save analysis: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	a.ID = id

	return nil
}

func (db *DB) GetAnalysisByTrader(traderID string) (*Analysis, error) {
	query := `SELECT id, trader_id, thesis, created_at FROM analyses
			  WHERE trader_id = ? ORDER BY created_at DESC LIMIT 1`
	row := db.conn.QueryRow(query, traderID)

	var a Analysis
	err := row.Scan(&a.ID, &a.TraderID, &a.Thesis, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get analysis: %w", err)
	}
	return &a, nil
}

func (db *DB) GetAllAnalysesByTrader(traderID string) ([]Analysis, error) {
	query := `SELECT id, trader_id, thesis, created_at FROM analyses
			  WHERE trader_id = ? ORDER BY created_at DESC`
	rows, err := db.conn.Query(query, traderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get analyses: %w", err)
	}
	defer rows.Close()

	var analyses []Analysis
	for rows.Next() {
		var a Analysis
		if err := rows.Scan(&a.ID, &a.TraderID, &a.Thesis, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan analysis: %w", err)
		}
		analyses = append(analyses, a)
	}
	return analyses, nil
}

func (db *DB) DeleteAnalysis(id int64) error {
	query := `DELETE FROM analyses WHERE id = ?`
	_, err := db.conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete analysis: %w", err)
	}
	return nil
}
