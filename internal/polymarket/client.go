package polymarket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/time/rate"
)

const (
	DefaultGammaBaseURL = "https://gamma-api.polymarket.com"
	DefaultCLOBBaseURL  = "https://clob.polymarket.com"
)

type Client struct {
	gammaResty  *resty.Client
	clobResty   *resty.Client
	rateLimiter *rate.Limiter
}

type Config struct {
	GammaBaseURL string
	CLOBBaseURL  string
	APIKey       string
	APISecret    string
	Passphrase   string
	Timeout      time.Duration
	RateLimit    rate.Limit
	Burst        int
}

func NewClient(cfg Config) *Client {
	if cfg.GammaBaseURL == "" {
		cfg.GammaBaseURL = DefaultGammaBaseURL
	}
	if cfg.CLOBBaseURL == "" {
		cfg.CLOBBaseURL = DefaultCLOBBaseURL
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	if cfg.RateLimit == 0 {
		cfg.RateLimit = rate.Every(time.Second / 5) // 5 requests per second default
	}
	if cfg.Burst == 0 {
		cfg.Burst = 10
	}

	limiter := rate.NewLimiter(cfg.RateLimit, cfg.Burst)

	gammaResty := resty.New().
		SetBaseURL(cfg.GammaBaseURL).
		SetTimeout(cfg.Timeout).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second)

	clobResty := resty.New().
		SetBaseURL(cfg.CLOBBaseURL).
		SetTimeout(cfg.Timeout).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second)

	c := &Client{
		gammaResty:  gammaResty,
		clobResty:   clobResty,
		rateLimiter: limiter,
	}

	// Apply rate limiting middleware
	gammaResty.OnBeforeRequest(c.beforeRequest)
	clobResty.OnBeforeRequest(c.beforeRequest)

	return c
}

func (c *Client) beforeRequest(client *resty.Client, req *resty.Request) error {
	ctx := req.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	return c.rateLimiter.Wait(ctx)
}

// ErrorResponse represents a generic API error
type ErrorResponse struct {
	Error string `json:"error"`
}

func (c *Client) checkError(resp *resty.Response) error {
	if resp.IsError() {
		if resp.StatusCode() == http.StatusTooManyRequests {
			return fmt.Errorf("rate limited: %s", resp.Status())
		}
		var errResp ErrorResponse
		if err := json.Unmarshal(resp.Body(), &errResp); err == nil && errResp.Error != "" {
			return fmt.Errorf("api error: %s (status: %s)", errResp.Error, resp.Status())
		}
		return fmt.Errorf("http error: %s", resp.Status())
	}
	return nil
}
