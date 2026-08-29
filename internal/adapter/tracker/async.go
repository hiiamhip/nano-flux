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