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
