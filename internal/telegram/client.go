package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Client struct {
	token   string
	chatID  string
	baseURL string
	http    *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token:   token,
		baseURL: fmt.Sprintf("https://api.telegram.org/bot%s", token),
		http:    &http.Client{},
	}
}

type sendMessageRequest struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

type getUpdatesResponse struct {
	OK     bool     `json:"ok"`
	Result []update `json:"result"`
}

type update struct {
	UpdateID int     `json:"update_id"`
	Message  *message `json:"message"`
}

type message struct {
	Chat chat   `json:"chat"`
	Text string `json:"text"`
}

type chat struct {
	ID int64 `json:"id"`
}

func (c *Client) WaitForStart(ctx context.Context) error {
	log.Println("waiting for /start message in Telegram...")
	offset := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		url := fmt.Sprintf("%s/getUpdates?offset=%d&timeout=10", c.baseURL, offset)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}

		resp, err := c.http.Do(req)
		if err != nil {
			log.Printf("getUpdates failed: %v, retrying...", err)
			time.Sleep(2 * time.Second)
			continue
		}

		var result getUpdatesResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()
		if err != nil {
			log.Printf("decode failed: %v, retrying...", err)
			time.Sleep(2 * time.Second)
			continue
		}

		for _, u := range result.Result {
			offset = u.UpdateID + 1
			if u.Message != nil && u.Message.Text == "/start" {
				c.chatID = strconv.FormatInt(u.Message.Chat.ID, 10)
				log.Printf("received /start from chat %s", c.chatID)
				return nil
			}
		}
	}
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
