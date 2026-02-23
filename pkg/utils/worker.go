package utils

import "sync"

// WaitGroup dùng để đếm số lượng các tiến trình ngầm đang chạy (như gửi mail)
var WorkerGroup sync.WaitGroup
var semaphore = make(chan struct{}, 20)

// RunInBackground chạy một hàm mà không làm chậm tốc độ trả response cho Client
func RunInBackground(fn func()) {
	semaphore <- struct{}{}
	WorkerGroup.Add(1)

	go func() {
		defer WorkerGroup.Done()
		defer func() { <-semaphore }()
		fn()
	}()
}

// Wait sẽ chặn API dừng lại cho đến khi tất cả các tiến trình ngầm hoàn tất
func Wait() {
	WorkerGroup.Wait()
}
