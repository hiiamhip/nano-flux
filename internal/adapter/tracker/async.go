package tracker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/hiiamhip/nanoflux/internal/domain"
)

type BatchSaver interface {
    SaveBatch(ctx context.Context, clicks []*domain.Click) error
}

type AsyncTracker struct {
    saver         BatchSaver         // Nơi để gửi mảng batch xuống (PostgreSQL repo)
    events        chan *domain.Click // Hàng đợi trong RAM chứa các click lẻ
    batchSize     int                // Số lượng gom tối đa (vd: 50)
    flushInterval time.Duration      // Thời gian gom tối đa (vd: 1 giây)
    wg            sync.WaitGroup     // Quản lý dừng worker an toàn
}

func NewAsyncTracker(saver BatchSaver, bufferCap int, batchSize int, flushInterval time.Duration) *AsyncTracker {
    t := &AsyncTracker{
        saver:         saver,
        events:        make(chan *domain.Click, bufferCap),
        batchSize:     batchSize,
        flushInterval: flushInterval,
    }
    t.wg.Add(1)
    go t.worker()
    return t
}

func (t *AsyncTracker) worker() {
    defer t.wg.Done()

    ticker := time.NewTicker(t.flushInterval)
    defer ticker.Stop()

    batch := make([]*domain.Click, 0, t.batchSize)

    for {
        select {
        case click, ok := <-t.events:
            if !ok {
                if len(batch) > 0 {
                    t.flush(batch)
                }
                return
            }

            batch = append(batch, click)
            if len(batch) >= t.batchSize {
                t.flush(batch)
                batch = make([]*domain.Click, 0, t.batchSize)
            }

        case <-ticker.C:
            if len(batch) > 0 {
                t.flush(batch)
                batch = make([]*domain.Click, 0, t.batchSize)
            }
        }
    }
}

func (t *AsyncTracker) flush(batch []*domain.Click) {
    ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
    defer cancel()

    if err := t.saver.SaveBatch(ctx, batch); err != nil {
        log.Printf("async tracker flush error: %v", err)
    }
}

func (t *AsyncTracker) Track(ctx context.Context, click *domain.Click) error {
    select {
    case t.events <- click:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    default:
        return nil
    }
}


func (t *AsyncTracker) Close() {
    close(t.events)
    t.wg.Wait()
}