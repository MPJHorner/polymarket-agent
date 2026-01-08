package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

func NewDB(path string) (*DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	instance := &DB{conn: db}
	if err := instance.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return instance, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS traders (
			address TEXT PRIMARY KEY,
			username TEXT,
			win_rate REAL,
			profit_loss REAL,
			roi REAL,
			volume REAL,
			last_scanned DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS trades (
			id TEXT PRIMARY KEY,
			trader_id TEXT,
			market_id TEXT,
			type TEXT,
			side TEXT,
			price REAL,
			size REAL,
			timestamp DATETIME,
			FOREIGN KEY(trader_id) REFERENCES traders(address)
		)`,
		`CREATE TABLE IF NOT EXISTS markets (
			id TEXT PRIMARY KEY,
			question TEXT,
			description TEXT,
			category TEXT,
			ends_at DATETIME,
			status TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS market_snapshots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			market_id TEXT,
			yes_price REAL,
			no_price REAL,
			timestamp DATETIME,
			FOREIGN KEY(market_id) REFERENCES markets(id)
		)`,
		`CREATE TABLE IF NOT EXISTS analyses (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			trader_id TEXT,
			thesis TEXT,
			created_at DATETIME,
			FOREIGN KEY(trader_id) REFERENCES traders(address)
		)`,
		`CREATE TABLE IF NOT EXISTS watchlist (
			trader_id TEXT PRIMARY KEY,
			notes TEXT,
			created_at DATETIME,
			FOREIGN KEY(trader_id) REFERENCES traders(address)
		)`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT
		)`,
	}

	for _, q := range queries {
		if _, err := db.conn.Exec(q); err != nil {
			return fmt.Errorf("failed to execute migration query (%s): %w", q, err)
		}
	}

	return nil
}
