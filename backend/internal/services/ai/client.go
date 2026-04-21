package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/liteoj/liteoj/backend/internal/config"
)

// Client is an OpenAI-compatible chat completions client suitable for Bifrost.
// Phase 1 ships this as a thin stub so handlers can call it; Phase 2 wires the
// three prompt flows (wrong-answer analysis, tag generation, solution writing).
//
// Timeout policy: the HTTP client itself has NO Timeout — request lifetime is
// driven entirely by the per-request context (handlers use 60s..180s). A
// hardcoded client.Timeout would override the context and cut the body read
// mid-stream, which previously surfaced as a useless "unexpected end of JSON
// input" error from the parser.
type Client struct {
	BaseURL string
	APIKey  string
	Model   string
	HTTP    *http.Client
}

func NewFromConfig(c *config.Config) *Client {
	return &Client{
		BaseURL: c.BifrostBaseURL,
		APIKey:  c.BifrostAPIKey,
		Model:   c.BifrostModel,
		HTTP:    &http.Client{},
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatReq struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type chatResp struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

// Chat returns the assistant text plus the raw HTTP body so callers can log
// what the upstream actually sent. raw is populated even on error paths
// (non-2xx, parse failure, body read failure) so the AI queue audit log can
// show the response that broke things.
func (c *Client) Chat(ctx context.Context, messages []Message) (string, string, error) {
	if c.BaseURL == "" || c.APIKey == "" {
		return "", "", errors.New("ai: BIFROST_BASE_URL / BIFROST_API_KEY not configured")
	}
	body, _ := json.Marshal(chatReq{Model: c.Model, Messages: messages})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	data, readErr := io.ReadAll(resp.Body)
	raw := string(data)
	if readErr != nil {
		// Surface the read error explicitly. Common cause: the request context
		// (or, previously, the client Timeout) fired during body streaming and
		// io.ReadAll returned partial bytes — silently dropping that error
		// hands a half-formed JSON to the parser and produces "unexpected end
		// of JSON input", masking the real cause.
		return "", raw, fmt.Errorf("ai: read response body: %w", readErr)
	}
	if resp.StatusCode/100 != 2 {
		return "", raw, fmt.Errorf("ai: %d %s", resp.StatusCode, raw)
	}
	var out chatResp
	if err := json.Unmarshal(data, &out); err != nil {
		return "", raw, fmt.Errorf("ai: parse response: %w", err)
	}
	if len(out.Choices) == 0 {
		return "", raw, errors.New("ai: empty response")
	}
	return out.Choices[0].Message.Content, raw, nil
}
