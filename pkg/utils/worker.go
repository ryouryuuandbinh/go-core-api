package utils

import "sync"

// WaitGroup dùng để đếm số lượng các tiến trình ngầm đang chạy (như gửi mail)
var WorkerGroup sync.WaitGroup

// RunInBackground chạy một hàm mà không làm chậm tốc độ trả response cho Client
func RunInBackground(fn func()) {
	WorkerGroup.Add(1) // Báo hiệu: "Tôi đang chạy"
	go func() {
		defer WorkerGroup.Done() // Báo hiệu: "Tôi xong rồi"
		fn()
	}()
}

// Wait sẽ chặn API dừng lại cho đến khi tất cả các tiến trình ngầm hoàn tất
func Wait() {
	WorkerGroup.Wait()
}
