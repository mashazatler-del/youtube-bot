package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	token   string
	chatID  string
	baseURL string
	http    *http.Client
}

func NewClient(token, chatID string) *Client {
	return &Client{
		token:   token,
		chatID:  chatID,
		baseURL: fmt.Sprintf("https://api.telegram.org/bot%s", token),
		http:    &http.Client{},
	}
}

type sendMessageRequest struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

func (c *Client) SendMessage(ctx context.Context, text string) error {
	body, err := json.Marshal(sendMessageRequest{
		ChatID: c.chatID,
		Text:   text,
	})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/sendMessage", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram api: status %d", resp.StatusCode)
	}

	return nil
}
