package tracker

import (
	"context"
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
    cancel        context.CancelFunc // Hàm kích hoạt dừng
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
                // Channel đã bị đóng và trong RAM đã hết sạch click!
                // Flush mẻ cuối cùng còn sót trong mảng batch rồi thoát
                if len(batch) > 0 {
                    // flush batch
                }
                return
            }

            batch = append(batch, click)
            if len(batch) >= t.batchSize {
                // flush batch
                batch = make([]*domain.Click, 0, t.batchSize)
            }

        case <-ticker.C:
            if len(batch) > 0 {
                // flush batch
                batch = make([]*domain.Click, 0, t.batchSize)
            }
        }
    }
}


func (t *AsyncTracker) Close() {
    close(t.events)
    t.wg.Wait()
}