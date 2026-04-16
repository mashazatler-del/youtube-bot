# YouTube Subscriber Notification Bot — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Go service that polls YouTube subscriber count every 30 minutes and sends Telegram notifications on change.

**Architecture:** Single-binary Go service with four packages — youtube client, telegram client, poller loop, and main entrypoint. No external dependencies, stdlib only. State kept in memory.

**Tech Stack:** Go (stdlib only), YouTube Data API v3, Telegram Bot API

---

## File Structure

```
youtube-bot/
├── cmd/bot/main.go              — entrypoint: config, wiring, graceful shutdown
├── internal/youtube/client.go   — YouTube API client
├── internal/youtube/client_test.go
├── internal/telegram/client.go  — Telegram Bot API client
├── internal/telegram/client_test.go
├── internal/poller/poller.go    — polling loop with change detection
├── internal/poller/poller_test.go
├── go.mod
└── .env.example                 — example env vars
```

---

### Task 1: Project Init

**Files:**
- Create: `go.mod`
- Create: `.env.example`

- [ ] **Step 1: Initialize Go module**

Run:
```bash
cd /Users/karpov/Documents/Projects/My/youtube-bot
go mod init youtube-bot
```

Expected: `go.mod` created with `module youtube-bot` and go version.

- [ ] **Step 2: Create .env.example**

Create `.env.example`:
```
YOUTUBE_API_KEY=your-youtube-api-key
YOUTUBE_CHANNEL_HANDLE=mashazatler
TELEGRAM_BOT_TOKEN=your-telegram-bot-token
TELEGRAM_CHAT_ID=your-chat-id
```

- [ ] **Step 3: Commit**

```bash
git init
git add go.mod .env.example
git commit -m "init: go module and env example"
```

---

### Task 2: YouTube Client

**Files:**
- Create: `internal/youtube/client.go`
- Create: `internal/youtube/client_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/youtube/client_test.go`:
```go
package youtube

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetSubscriberCount(t *testing.T) {
	resp := channelsResponse{
		Items: []channelItem{
			{Statistics: statistics{SubscriberCount: "12345"}},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("forHandle") != "testhandle" {
			t.Errorf("expected forHandle=testhandle, got %s", r.URL.Query().Get("forHandle"))
		}
		if r.URL.Query().Get("key") != "testkey" {
			t.Errorf("expected key=testkey, got %s", r.URL.Query().Get("key"))
		}
		if r.URL.Query().Get("part") != "statistics" {
			t.Errorf("expected part=statistics, got %s", r.URL.Query().Get("part"))
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := NewClient("testkey", "testhandle")
	c.baseURL = srv.URL

	count, err := c.GetSubscriberCount(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 12345 {
		t.Fatalf("expected 12345, got %d", count)
	}
}

func TestGetSubscriberCountNoItems(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(channelsResponse{Items: []channelItem{}})
	}))
	defer srv.Close()

	c := NewClient("testkey", "testhandle")
	c.baseURL = srv.URL

	_, err := c.GetSubscriberCount(context.Background())
	if err == nil {
		t.Fatal("expected error for empty items")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/karpov/Documents/Projects/My/youtube-bot && go test ./internal/youtube/ -v`

Expected: FAIL — package does not exist yet.

- [ ] **Step 3: Write implementation**

Create `internal/youtube/client.go`:
```go
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/karpov/Documents/Projects/My/youtube-bot && go test ./internal/youtube/ -v`

Expected: PASS (both tests).

- [ ] **Step 5: Commit**

```bash
git add internal/youtube/
git commit -m "feat: youtube api client with subscriber count"
```

---

### Task 3: Telegram Client

**Files:**
- Create: `internal/telegram/client.go`
- Create: `internal/telegram/client_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/telegram/client_test.go`:
```go
package telegram

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendMessage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var body sendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body.ChatID != "123" {
			t.Errorf("expected chat_id=123, got %s", body.ChatID)
		}
		if body.Text != "hello" {
			t.Errorf("expected text=hello, got %s", body.Text)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}))
	defer srv.Close()

	c := NewClient("testtoken", "123")
	c.baseURL = srv.URL

	err := c.SendMessage(context.Background(), "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSendMessageError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	c := NewClient("testtoken", "123")
	c.baseURL = srv.URL

	err := c.SendMessage(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error for bad status")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/karpov/Documents/Projects/My/youtube-bot && go test ./internal/telegram/ -v`

Expected: FAIL — package does not exist yet.

- [ ] **Step 3: Write implementation**

Create `internal/telegram/client.go`:
```go
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/karpov/Documents/Projects/My/youtube-bot && go test ./internal/telegram/ -v`

Expected: PASS (both tests).

- [ ] **Step 5: Commit**

```bash
git add internal/telegram/
git commit -m "feat: telegram bot api client"
```

---

### Task 4: Poller

**Files:**
- Create: `internal/poller/poller.go`
- Create: `internal/poller/poller_test.go`

- [ ] **Step 1: Write the failing test**

Create `internal/poller/poller_test.go`:
```go
package poller

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

type fakeYouTube struct {
	counts []int64
	index  int
	mu     sync.Mutex
}

func (f *fakeYouTube) GetSubscriberCount(ctx context.Context) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.index >= len(f.counts) {
		return f.counts[len(f.counts)-1], nil
	}
	c := f.counts[f.index]
	f.index++
	return c, nil
}

type fakeTelegram struct {
	messages []string
	mu       sync.Mutex
}

func (f *fakeTelegram) SendMessage(ctx context.Context, text string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.messages = append(f.messages, text)
	return nil
}

func (f *fakeTelegram) getMessages() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := make([]string, len(f.messages))
	copy(cp, f.messages)
	return cp
}

func TestPollerNoNotificationOnFirstCheck(t *testing.T) {
	yt := &fakeYouTube{counts: []int64{100, 100}}
	tg := &fakeTelegram{}

	p := New(yt, tg, 50*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 130*time.Millisecond)
	defer cancel()

	p.Run(ctx)

	msgs := tg.getMessages()
	if len(msgs) != 0 {
		t.Fatalf("expected no messages, got %v", msgs)
	}
}

func TestPollerNotifiesOnIncrease(t *testing.T) {
	yt := &fakeYouTube{counts: []int64{100, 103}}
	tg := &fakeTelegram{}

	p := New(yt, tg, 50*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 130*time.Millisecond)
	defer cancel()

	p.Run(ctx)

	msgs := tg.getMessages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d: %v", len(msgs), msgs)
	}
	expected := fmt.Sprintf("📈 Подписчики: 100 → 103 (+3)")
	if msgs[0] != expected {
		t.Fatalf("expected %q, got %q", expected, msgs[0])
	}
}

func TestPollerNotifiesOnDecrease(t *testing.T) {
	yt := &fakeYouTube{counts: []int64{100, 97}}
	tg := &fakeTelegram{}

	p := New(yt, tg, 50*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 130*time.Millisecond)
	defer cancel()

	p.Run(ctx)

	msgs := tg.getMessages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d: %v", len(msgs), msgs)
	}
	expected := fmt.Sprintf("📉 Подписчики: 100 → 97 (-3)")
	if msgs[0] != expected {
		t.Fatalf("expected %q, got %q", expected, msgs[0])
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `cd /Users/karpov/Documents/Projects/My/youtube-bot && go test ./internal/poller/ -v`

Expected: FAIL — package does not exist yet.

- [ ] **Step 3: Write implementation**

Create `internal/poller/poller.go`:
```go
package poller

import (
	"context"
	"fmt"
	"log"
	"time"
)

type SubscriberCounter interface {
	GetSubscriberCount(ctx context.Context) (int64, error)
}

type MessageSender interface {
	SendMessage(ctx context.Context, text string) error
}

type Poller struct {
	youtube  SubscriberCounter
	telegram MessageSender
	interval time.Duration
}

func New(youtube SubscriberCounter, telegram MessageSender, interval time.Duration) *Poller {
	return &Poller{
		youtube:  youtube,
		telegram: telegram,
		interval: interval,
	}
}

func (p *Poller) Run(ctx context.Context) {
	count, err := p.youtube.GetSubscriberCount(ctx)
	if err != nil {
		log.Printf("initial check failed: %v", err)
		return
	}
	prev := count
	log.Printf("initial subscriber count: %d", prev)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			count, err := p.youtube.GetSubscriberCount(ctx)
			if err != nil {
				log.Printf("check failed: %v", err)
				continue
			}

			if count != prev {
				diff := count - prev
				msg := formatMessage(prev, count, diff)
				if err := p.telegram.SendMessage(ctx, msg); err != nil {
					log.Printf("send message failed: %v", err)
				}
				prev = count
			}
		}
	}
}

func formatMessage(prev, current, diff int64) string {
	if diff > 0 {
		return fmt.Sprintf("📈 Подписчики: %d → %d (+%d)", prev, current, diff)
	}
	return fmt.Sprintf("📉 Подписчики: %d → %d (%d)", prev, current, diff)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /Users/karpov/Documents/Projects/My/youtube-bot && go test ./internal/poller/ -v`

Expected: PASS (all three tests).

- [ ] **Step 5: Commit**

```bash
git add internal/poller/
git commit -m "feat: poller with change detection and notifications"
```

---

### Task 5: Main Entrypoint

**Files:**
- Create: `cmd/bot/main.go`

- [ ] **Step 1: Write main.go**

Create `cmd/bot/main.go`:
```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"youtube-bot/internal/poller"
	"youtube-bot/internal/telegram"
	"youtube-bot/internal/youtube"
)

func main() {
	apiKey := requireEnv("YOUTUBE_API_KEY")
	handle := requireEnv("YOUTUBE_CHANNEL_HANDLE")
	botToken := requireEnv("TELEGRAM_BOT_TOKEN")
	chatID := requireEnv("TELEGRAM_CHAT_ID")

	yt := youtube.NewClient(apiKey, handle)
	tg := telegram.NewClient(botToken, chatID)

	p := poller.New(yt, tg, 30*time.Minute)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		log.Printf("received signal %v, shutting down", sig)
		cancel()
	}()

	log.Println("bot started")
	p.Run(ctx)
	log.Println("bot stopped")
}

func requireEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("required env var %s is not set", key)
	}
	return val
}
```

- [ ] **Step 2: Verify it compiles**

Run: `cd /Users/karpov/Documents/Projects/My/youtube-bot && go build ./cmd/bot/`

Expected: builds successfully, creates `bot` binary.

- [ ] **Step 3: Commit**

```bash
git add cmd/bot/main.go
git commit -m "feat: main entrypoint with graceful shutdown"
```

---

### Task 6: Run All Tests and Manual Smoke Test

- [ ] **Step 1: Run all tests**

Run: `cd /Users/karpov/Documents/Projects/My/youtube-bot && go test ./... -v`

Expected: all tests PASS.

- [ ] **Step 2: Manual smoke test**

Run with real credentials (user provides values):
```bash
cd /Users/karpov/Documents/Projects/My/youtube-bot
YOUTUBE_API_KEY=<key> \
YOUTUBE_CHANNEL_HANDLE=mashazatler \
TELEGRAM_BOT_TOKEN=<token> \
TELEGRAM_CHAT_ID=<chat_id> \
go run ./cmd/bot/
```

Expected: logs `bot started` and `initial subscriber count: <number>`. Press Ctrl+C to stop — logs `bot stopped`.

- [ ] **Step 3: Final commit with .gitignore**

Create `.gitignore`:
```
bot
.env
.idea/
```

```bash
git add .gitignore
git commit -m "chore: add gitignore"
```
