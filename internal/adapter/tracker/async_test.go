package tracker

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/hiiamhip/nanoflux/internal/domain"
)

type mockBatchSaver struct {
	mu     sync.Mutex
	saved  [][]*domain.Click
	saveCh chan []*domain.Click
}

func newMockBatchSaver() *mockBatchSaver {
	return &mockBatchSaver{
		saved:  make([][]*domain.Click, 0),
		saveCh: make(chan []*domain.Click, 10),
	}
}

func (m *mockBatchSaver) SaveBatch(ctx context.Context, clicks []*domain.Click) error {
	m.mu.Lock()
	m.saved = append(m.saved, clicks)
	m.mu.Unlock()

	select {
	case m.saveCh <- clicks:
	default:
	}
	return nil
}

func TestAsyncTracker_BatchSizeFlush(t *testing.T) {
	saver := newMockBatchSaver()
	// bufferCap: 100, batchSize: 3, flushInterval: 10s (chỉ flush khi đủ batch 3)
	tracker := NewAsyncTracker(saver, 100, 3, 10*time.Second)
	defer tracker.Close()

	ctx := context.Background()

	_ = tracker.Track(ctx, &domain.Click{ID: "1", ShortCode: "a"})
	_ = tracker.Track(ctx, &domain.Click{ID: "2", ShortCode: "b"})
	_ = tracker.Track(ctx, &domain.Click{ID: "3", ShortCode: "c"})

	select {
	case batch := <-saver.saveCh:
		if len(batch) != 3 {
			t.Fatalf("expected batch size 3, got %d", len(batch))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for batch flush")
	}
}

func TestAsyncTracker_TickerFlush(t *testing.T) {
	saver := newMockBatchSaver()
	// bufferCap: 100, batchSize: 100 (rất lớn), flushInterval: 100ms
	tracker := NewAsyncTracker(saver, 100, 100, 100*time.Millisecond)
	defer tracker.Close()

	ctx := context.Background()
	_ = tracker.Track(ctx, &domain.Click{ID: "1", ShortCode: "a"})

	select {
	case batch := <-saver.saveCh:
		if len(batch) != 1 {
			t.Fatalf("expected batch size 1 from ticker flush, got %d", len(batch))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for ticker flush")
	}
}

func TestAsyncTracker_CloseFlushesRemaining(t *testing.T) {
	saver := newMockBatchSaver()
	tracker := NewAsyncTracker(saver, 100, 100, 10*time.Second)

	ctx := context.Background()
	_ = tracker.Track(ctx, &domain.Click{ID: "1", ShortCode: "a"})
	_ = tracker.Track(ctx, &domain.Click{ID: "2", ShortCode: "b"})

	tracker.Close()

	saver.mu.Lock()
	defer saver.mu.Unlock()

	totalSaved := 0
	for _, batch := range saver.saved {
		totalSaved += len(batch)
	}

	if totalSaved != 2 {
		t.Fatalf("expected 2 clicks flushed on close, got %d", totalSaved)
	}
}
