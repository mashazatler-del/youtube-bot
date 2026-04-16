package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type Client struct {
	apiKey  string
	handle  string
	baseURL string
	http    *http.Client
}

func NewClient(apiKey, handle string) *Client {
	return &Client{
		apiKey:  apiKey,
		handle:  handle,
		baseURL: "https://www.googleapis.com/youtube/v3",
		http:    &http.Client{},
	}
}

type channelsResponse struct {
	Items []channelItem `json:"items"`
}

type channelItem struct {
	Statistics statistics `json:"statistics"`
}

type statistics struct {
	SubscriberCount string `json:"subscriberCount"`
}

func (c *Client) GetSubscriberCount(ctx context.Context) (int64, error) {
	url := fmt.Sprintf("%s/channels?part=statistics&forHandle=%s&key=%s", c.baseURL, c.handle, c.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("youtube api: status %d", resp.StatusCode)
	}

	var result channelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode response: %w", err)
	}

	if len(result.Items) == 0 {
		return 0, fmt.Errorf("channel not found for handle %q", c.handle)
	}

	count, err := strconv.ParseInt(result.Items[0].Statistics.SubscriberCount, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse subscriber count: %w", err)
	}

	return count, nil
}
