package db

import (
	"time"
)

type Trader struct {
	Address     string    `json:"address"`
	Username    string    `json:"username"`
	WinRate     float64   `json:"win_rate"`
	ProfitLoss  float64   `json:"profit_loss"`
	ROI         float64   `json:"roi"`
	Volume      float64   `json:"volume"`
	LastScanned time.Time `json:"last_scanned"`
}

type Trade struct {
	ID        string    `json:"id"`
	TraderID  string    `json:"trader_id"`
	MarketID  string    `json:"market_id"`
	Type      string    `json:"type"` // buy/sell
	Side      string    `json:"side"` // yes/no
	Price     float64   `json:"price"`
	Size      float64   `json:"size"`
	Timestamp time.Time `json:"timestamp"`
}

type Market struct {
	ID          string    `json:"id"`
	Question    string    `json:"question"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	EndsAt      time.Time `json:"ends_at"`
	Status      string    `json:"status"` // open/closed/resolved
}

type MarketSnapshot struct {
	ID        int64     `json:"id"`
	MarketID  string    `json:"market_id"`
	YesPrice  float64   `json:"yes_price"`
	NoPrice   float64   `json:"no_price"`
	Timestamp time.Time `json:"timestamp"`
}

type Analysis struct {
	ID        int64     `json:"id"`
	TraderID  string    `json:"trader_id"`
	Thesis    string    `json:"thesis"`
	CreatedAt time.Time `json:"created_at"`
}

type WatchlistItem struct {
	TraderID  string    `json:"trader_id"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
}

type Setting struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
