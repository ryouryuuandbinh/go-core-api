package utils

import "sync"

// WaitGroup dùng để đếm số lượng các tiến trình ngầm đang chạy (như gửi mail)
var WorkerGroup sync.WaitGroup
var semaphore = make(chan struct{}, 20)

// RunInBackground chạy một hàm mà không làm chậm tốc độ trả response cho Client
func RunInBackground(fn func()) {
	WorkerGroup.Add(1)
	go func() {
		defer WorkerGroup.Done()

		semaphore <- struct{}{}        // Xin 1 slot
		defer func() { <-semaphore }() // Trả slot khi xong việc

		fn()
	}()
}

// Wait sẽ chặn API dừng lại cho đến khi tất cả các tiến trình ngầm hoàn tất
func Wait() {
	WorkerGroup.Wait()
}
