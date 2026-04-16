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
