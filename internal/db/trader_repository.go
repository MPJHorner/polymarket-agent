package db

import (
	"database/sql"
	"fmt"
)

func (db *DB) SaveTrader(t *Trader) error {
	query := `INSERT INTO traders (address, username, win_rate, profit_loss, roi, volume, last_scanned)
			  VALUES (?, ?, ?, ?, ?, ?, ?)
			  ON CONFLICT(address) DO UPDATE SET
			  username=excluded.username,
			  win_rate=excluded.win_rate,
			  profit_loss=excluded.profit_loss,
			  roi=excluded.roi,
			  volume=excluded.volume,
			  last_scanned=excluded.last_scanned`
	
	_, err := db.conn.Exec(query, t.Address, t.Username, t.WinRate, t.ProfitLoss, t.ROI, t.Volume, t.LastScanned)
	if err != nil {
		return fmt.Errorf("failed to save trader: %w", err)
	}
	return nil
}

func (db *DB) GetTrader(address string) (*Trader, error) {
	query := `SELECT address, username, win_rate, profit_loss, roi, volume, last_scanned FROM traders WHERE address = ?`
	row := db.conn.QueryRow(query, address)

	var t Trader
	err := row.Scan(&t.Address, &t.Username, &t.WinRate, &t.ProfitLoss, &t.ROI, &t.Volume, &t.LastScanned)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get trader: %w", err)
	}
	return &t, nil
}

type SortField string

const (
	SortByProfitLoss SortField = "profit_loss"
	SortByWinRate    SortField = "win_rate"
	SortByROI        SortField = "roi"
	SortByVolume     SortField = "volume"
)

type SortOrder string

const (
	SortAsc  SortOrder = "ASC"
	SortDesc SortOrder = "DESC"
)

type ListTradersOptions struct {
	SortBy    SortField
	Order     SortOrder
	Limit     int
	Offset    int
}

func (db *DB) ListTraders() ([]Trader, error) {
	return db.ListTradersWithOptions(ListTradersOptions{
		SortBy: SortByProfitLoss,
		Order:  SortDesc,
	})
}

func (db *DB) ListTradersWithOptions(opts ListTradersOptions) ([]Trader, error) {
	if opts.SortBy == "" {
		opts.SortBy = SortByProfitLoss
	}
	if opts.Order == "" {
		opts.Order = SortDesc
	}

	query := fmt.Sprintf(`SELECT address, username, win_rate, profit_loss, roi, volume, last_scanned FROM traders ORDER BY %s %s`, opts.SortBy, opts.Order)

	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", opts.Limit)
		if opts.Offset > 0 {
			query += fmt.Sprintf(" OFFSET %d", opts.Offset)
		}
	}

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list traders: %w", err)
	}
	defer rows.Close()

	var traders []Trader
	for rows.Next() {
		var t Trader
		if err := rows.Scan(&t.Address, &t.Username, &t.WinRate, &t.ProfitLoss, &t.ROI, &t.Volume, &t.LastScanned); err != nil {
			return nil, fmt.Errorf("failed to scan trader: %w", err)
		}
		traders = append(traders, t)
	}
	return traders, nil
}

func (db *DB) CountTraders() (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM traders").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count traders: %w", err)
	}
	return count, nil
}
