package polymarket

import (
	"context"
	"fmt"
)

type Market struct {
	ID            string   `json:"id"`
	Question      string   `json:"question"`
	ConditionID   string   `json:"conditionId"`
	Slug          string   `json:"slug"`
	Resolution    string   `json:"resolution"`
	Tokens        []Token  `json:"tokens"`
	Active        bool     `json:"active"`
	Closed        bool     `json:"closed"`
	Archived      bool     `json:"archived"`
	Liquidity     float64  `json:"liquidity,string"`
	Volume        float64  `json:"volume,string"`
	OutcomeAssets []string `json:"outcomeAssets"`
}

type Token struct {
	TokenID string `json:"tokenId"`
	Outcome string `json:"outcome"`
	Price   float64 `json:"price"`
}

func (c *Client) GetMarket(ctx context.Context, id string) (*Market, error) {
	var market Market
	resp, err := c.gammaResty.R().
		SetContext(ctx).
		SetResult(&market).
		Get(fmt.Sprintf("/markets/%s", id))

	if err != nil {
		return nil, fmt.Errorf("failed to get market: %w", err)
	}

	if err := c.checkError(resp); err != nil {
		return nil, err
	}

	return &market, nil
}

func (c *Client) ListMarkets(ctx context.Context, limit int) ([]Market, error) {
	var markets []Market
	resp, err := c.gammaResty.R().
		SetContext(ctx).
		SetQueryParam("limit", fmt.Sprintf("%d", limit)).
		SetResult(&markets).
		Get("/markets")

	if err != nil {
		return nil, fmt.Errorf("failed to list markets: %w", err)
	}

	if err := c.checkError(resp); err != nil {
		return nil, err
	}

	return markets, nil
}
