package claude

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"polytracker/internal/db"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

var (
	ErrNoAPIKey       = errors.New("claude API key is not configured")
	ErrEmptyResponse  = errors.New("received empty response from Claude")
	ErrTokenLimit     = errors.New("response was truncated due to token limit")
	ErrInvalidTrader  = errors.New("trader data is invalid or missing")
)

// Config holds the configuration for the Claude client.
type Config struct {
	APIKey   string
	Endpoint string
}

// Client wraps the Anthropic SDK client.
type Client struct {
	client *anthropic.Client
	config Config
}

// NewClient creates a new Claude client with the given configuration.
func NewClient(cfg Config) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, ErrNoAPIKey
	}

	opts := []option.RequestOption{
		option.WithAPIKey(cfg.APIKey),
	}

	if cfg.Endpoint != "" {
		opts = append(opts, option.WithBaseURL(cfg.Endpoint))
	}

	client := anthropic.NewClient(opts...)

	return &Client{
		client: &client,
		config: cfg,
	}, nil
}

// TraderData contains all the data needed to generate a thesis prompt.
type TraderData struct {
	Trader  *db.Trader
	Trades  []db.Trade
	Markets map[string]*db.Market
}

// AnalysisResult contains the result of a trader analysis.
type AnalysisResult struct {
	Thesis      string
	Model       string
	InputTokens int64
	OutputTokens int64
	StopReason  string
	CreatedAt   time.Time
}

// AnalyzeTrader generates a trading thesis for the given trader using Claude.
func (c *Client) AnalyzeTrader(ctx context.Context, data TraderData) (*AnalysisResult, error) {
	if data.Trader == nil {
		return nil, ErrInvalidTrader
	}

	prompt := GenerateThesisPrompt(data)

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model("claude-sonnet-4-20250514"),
		MaxTokens: 4096,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	}

	resp, err := c.client.Messages.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	// Extract text content from response
	var thesis strings.Builder
	for _, block := range resp.Content {
		if block.Type == "text" {
			thesis.WriteString(block.Text)
		}
	}

	if thesis.Len() == 0 {
		return nil, ErrEmptyResponse
	}

	result := &AnalysisResult{
		Thesis:       thesis.String(),
		Model:        string(resp.Model),
		InputTokens:  resp.Usage.InputTokens,
		OutputTokens: resp.Usage.OutputTokens,
		StopReason:   string(resp.StopReason),
		CreatedAt:    time.Now(),
	}

	// Check if response was truncated
	if resp.StopReason == anthropic.StopReasonMaxTokens {
		return result, ErrTokenLimit
	}

	return result, nil
}

// GenerateThesisPrompt creates a structured prompt for Claude to analyze a trader.
func GenerateThesisPrompt(data TraderData) string {
	var sb strings.Builder

	sb.WriteString("You are an expert crypto trading analyst specializing in prediction markets. ")
	sb.WriteString("Analyze the following trader's activity on Polymarket and generate a detailed trading thesis.\n\n")

	// Trader profile section
	sb.WriteString("## Trader Profile\n\n")
	sb.WriteString(fmt.Sprintf("- **Address:** %s\n", data.Trader.Address))
	if data.Trader.Username != "" {
		sb.WriteString(fmt.Sprintf("- **Username:** %s\n", data.Trader.Username))
	}
	sb.WriteString(fmt.Sprintf("- **Win Rate:** %.2f%%\n", data.Trader.WinRate*100))
	sb.WriteString(fmt.Sprintf("- **Profit/Loss:** $%.2f\n", data.Trader.ProfitLoss))
	sb.WriteString(fmt.Sprintf("- **ROI:** %.2f%%\n", data.Trader.ROI*100))
	sb.WriteString(fmt.Sprintf("- **Total Volume:** $%.2f\n", data.Trader.Volume))
	sb.WriteString(fmt.Sprintf("- **Last Scanned:** %s\n\n", data.Trader.LastScanned.Format("2006-01-02 15:04:05")))

	// Trading history section
	sb.WriteString("## Recent Trading Activity\n\n")

	if len(data.Trades) == 0 {
		sb.WriteString("No trades available for analysis.\n\n")
	} else {
		// Limit to most recent 50 trades for prompt size management
		tradesToShow := data.Trades
		if len(tradesToShow) > 50 {
			tradesToShow = tradesToShow[:50]
		}

		for i, trade := range tradesToShow {
			marketQuestion := "Unknown Market"
			if market, ok := data.Markets[trade.MarketID]; ok && market != nil {
				marketQuestion = market.Question
			}

			sb.WriteString(fmt.Sprintf("### Trade %d\n", i+1))
			sb.WriteString(fmt.Sprintf("- **Market:** %s\n", marketQuestion))
			sb.WriteString(fmt.Sprintf("- **Type:** %s\n", trade.Type))
			sb.WriteString(fmt.Sprintf("- **Side:** %s\n", trade.Side))
			sb.WriteString(fmt.Sprintf("- **Price:** $%.4f\n", trade.Price))
			sb.WriteString(fmt.Sprintf("- **Size:** %.4f\n", trade.Size))
			sb.WriteString(fmt.Sprintf("- **Time:** %s\n\n", trade.Timestamp.Format("2006-01-02 15:04:05")))
		}

		if len(data.Trades) > 50 {
			sb.WriteString(fmt.Sprintf("_(Showing 50 of %d total trades)_\n\n", len(data.Trades)))
		}
	}

	// Analysis request
	sb.WriteString("## Analysis Request\n\n")
	sb.WriteString("Based on the trader profile and trading history above, please provide:\n\n")
	sb.WriteString("1. **Trading Strategy Summary:** What patterns do you observe in their trading behavior?\n")
	sb.WriteString("2. **Market Focus:** What types of markets or events does this trader focus on?\n")
	sb.WriteString("3. **Risk Profile:** How would you characterize their risk tolerance and position sizing?\n")
	sb.WriteString("4. **Timing Analysis:** Do they tend to trade early in market lifecycles or closer to resolution?\n")
	sb.WriteString("5. **Strengths:** What appears to be working well in their strategy?\n")
	sb.WriteString("6. **Weaknesses/Risks:** What potential weaknesses or risks do you identify?\n")
	sb.WriteString("7. **Overall Thesis:** A concise thesis statement summarizing this trader's approach and edge.\n\n")
	sb.WriteString("Please format your response in clear markdown sections.")

	return sb.String()
}
