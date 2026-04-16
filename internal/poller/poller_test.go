package poller

import (
	"context"
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
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message (start only), got %d: %v", len(msgs), msgs)
	}
	expected := "🚀 Бот запущен! Подписчиков: 100"
	if msgs[0] != expected {
		t.Fatalf("expected %q, got %q", expected, msgs[0])
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
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d: %v", len(msgs), msgs)
	}
	expected := "📈 Подписчики: 100 → 103 (+3)"
	if msgs[1] != expected {
		t.Fatalf("expected %q, got %q", expected, msgs[1])
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
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d: %v", len(msgs), msgs)
	}
	expected := "📉 Подписчики: 100 → 97 (-3)"
	if msgs[1] != expected {
		t.Fatalf("expected %q, got %q", expected, msgs[1])
	}
}
