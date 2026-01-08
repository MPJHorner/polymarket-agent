package polymarket

import (
	"context"
	"fmt"
)

type Trade struct {
	ID        string  `json:"id"`
	MarketID  string  `json:"market_id"`
	Price     float64 `json:"price,string"`
	Size      float64 `json:"size,string"`
	Side      string  `json:"side"`
	Timestamp int64   `json:"timestamp,string"`
	Maker     string  `json:"maker"`
	Taker     string  `json:"taker"`
}

type Orderbook struct {
	MarketID string `json:"market_id"`
	Asks     []Level `json:"asks"`
	Bids     []Level `json:"bids"`
}

type Level struct {
	Price float64 `json:"price,string"`
	Size  float64 `json:"size,string"`
}

func (c *Client) GetTrades(ctx context.Context, marketID string) ([]Trade, error) {
	var trades []Trade
	resp, err := c.clobResty.R().
		SetContext(ctx).
		SetQueryParam("market_id", marketID).
		SetResult(&trades).
		Get("/trades")

	if err != nil {
		return nil, fmt.Errorf("failed to get trades: %w", err)
	}

	if err := c.checkError(resp); err != nil {
		return nil, err
	}

	return trades, nil
}

func (c *Client) GetAccountTrades(ctx context.Context, address string) ([]Trade, error) {
	var trades []Trade
	resp, err := c.clobResty.R().
		SetContext(ctx).
		SetQueryParam("maker_address", address).
		SetResult(&trades).
		Get("/trades")

	if err != nil {
		return nil, fmt.Errorf("failed to get account trades: %w", err)
	}

	if err := c.checkError(resp); err != nil {
		return nil, err
	}

	return trades, nil
}

func (c *Client) GetOrderbook(ctx context.Context, tokenID string) (*Orderbook, error) {
	var book Orderbook
	resp, err := c.clobResty.R().
		SetContext(ctx).
		SetQueryParam("token_id", tokenID).
		SetResult(&book).
		Get("/book")

	if err != nil {
		return nil, fmt.Errorf("failed to get orderbook: %w", err)
	}

	if err := c.checkError(resp); err != nil {
		return nil, err
	}

	return &book, nil
}
