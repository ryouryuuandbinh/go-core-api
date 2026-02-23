package utils

import (
	"context"
	"go-core-api/pkg/logger"
	"sync"

	"go.uber.org/zap"
)

var (
	WorkerGroup sync.WaitGroup
	jobQueue    chan func()
	once        sync.Once
)

// InitWorkerPool khởi tạo số lượng Worker cố định (VD: 20 workers)
// Khai báo hàm này trong main.go: `utils.InitWorkerPool(ctx, 20)`
func InitWorkerPool(ctx context.Context, maxWorkers int) {
	once.Do(func() {
		jobQueue = make(chan func(), 1000) // Hàng đợi chứa tối đa 1000 jobs

		for i := 0; i < maxWorkers; i++ {
			WorkerGroup.Add(1)
			go worker(ctx)
		}
	})
}

// worker là các tiến trình chạy ngầm, luôn chờ lấy job ra để làm
func worker(ctx context.Context) {
	defer WorkerGroup.Done()
	for {
		select {
		case <-ctx.Done():
			// Server tắt -> Worker dừng
			return
		case job := <-jobQueue:
			executeJobSafe(job)
		}
	}
}

func executeJobSafe(job func()) {
	// Bắt lỗi Panic để không làm chết Worker
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic trong Background Worker", zap.Any("error", r))
		}
	}()
	job()
}

// RunInBackground đẩy job vào Pool ngay lập tức (Không tốn chi phí tạo Goroutine mới)
func RunInBackground(fn func()) {
	// Tránh trường hợp tràn Queue làm treo request
	select {
	case jobQueue <- fn:
		// Thành công đẩy job
	default:
		logger.Error("Job Queue đã đầy, bỏ qua tác vụ ngầm để bảo vệ Server")
	}
}
