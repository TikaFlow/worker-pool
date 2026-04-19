package pool

import (
	"sync"
	"testing"
	"time"
)

// TestNew 测试 New 函数
func TestNew(t *testing.T) {
	p := New(5, nil)
	if p == nil {
		t.Fatal("New 返回 nil")
	}
	p.CloseAndWait()
}

// TestNewWithZero 测试 workerCount 为 0 时自动修正
func TestNewWithZero(t *testing.T) {
	p := New(0, nil)
	if p == nil {
		t.Fatal("New 返回 nil")
	}
	p.CloseAndWait()
}

// TestNewWithNegative 测试 workerCount 为负数时自动修正
func TestNewWithNegative(t *testing.T) {
	p := New(-10, nil)
	if p == nil {
		t.Fatal("New 返回 nil")
	}
	p.CloseAndWait()
}

// TestAdd 测试添加任务
func TestAdd(t *testing.T) {
	var wg sync.WaitGroup
	p := New(4, nil)
	counter := 0
	var mu sync.Mutex

	for i := 0; i < 100; i++ {
		wg.Add(1)
		p.Add(func() {
			mu.Lock()
			counter++
			mu.Unlock()
			wg.Done()
		})
	}

	wg.Wait()
	p.CloseAndWait()

	if counter != 100 {
		t.Errorf("期望 100，实际 %d", counter)
	}
}

// TestAddNil 测试添加 nil 任务
func TestAddNil(t *testing.T) {
	p := New(4, nil)
	p.Add(nil) // 不应该 panic
	p.CloseAndWait()
}

// TestPanicHandler 测试 panic 处理
func TestPanicHandler(t *testing.T) {
	var panicCaught any
	var wg sync.WaitGroup
	p := New(4, &Config{
		PanicHandler: func(r any) {
			panicCaught = r
			wg.Done()
		},
	})

	wg.Add(1)
	p.Add(func() {
		panic("test panic")
	})

	wg.Wait()
	p.CloseAndWait()

	if panicCaught != "test panic" {
		t.Errorf("期望 'test panic'，实际 %v", panicCaught)
	}
}

// TestCloseMultipleTimes 测试多次调用 Close
func TestCloseMultipleTimes(t *testing.T) {
	p := New(4, nil)
	p.Close()
	p.Close() // 不应该 panic
	p.Close() // 不应该 panic
	p.CloseAndWait()
}

// TestCloseAndWait 测试 CloseAndWait
func TestCloseAndWait(t *testing.T) {
	p := New(4, nil)
	taskCount := 100
	var counter int
	var mu sync.Mutex

	for i := 0; i < taskCount; i++ {
		p.Add(func() {
			time.Sleep(1 * time.Millisecond)
			mu.Lock()
			counter++
			mu.Unlock()
		})
	}

	p.CloseAndWait()

	if counter != taskCount {
		t.Errorf("期望 %d，实际 %d", taskCount, counter)
	}
}

// TestAddAfterClose 测试关闭后添加任务
func TestAddAfterClose(t *testing.T) {
	p := New(4, nil)
	p.Close()
	time.Sleep(10 * time.Millisecond)
	// 不应该 panic
	p.Add(func() {})
	p.CloseAndWait()
}

// TestConcurrent 测试大量并发任务
func TestConcurrent(t *testing.T) {
	p := New(16, nil)
	taskCount := 10000
	var wg sync.WaitGroup

	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		p.Add(func() {
			// 模拟一些工作
			time.Sleep(50 * time.Microsecond)
			wg.Done()
		})
	}

	wg.Wait()
	p.CloseAndWait()
}

// TestWorkerCount 测试不同的 worker 数量
func TestWorkerCount(t *testing.T) {
	testCases := []int{1, 2, 4, 8, 16, 32}

	for _, wc := range testCases {
		t.Run(string(rune('0'+wc)), func(t *testing.T) {
			p := New(wc, nil)
			var wg sync.WaitGroup
			counter := 0
			var mu sync.Mutex

			for i := 0; i < 100; i++ {
				wg.Add(1)
				p.Add(func() {
					mu.Lock()
					counter++
					mu.Unlock()
					wg.Done()
				})
			}

			wg.Wait()
			p.CloseAndWait()

			if counter != 100 {
				t.Errorf("期望 100，实际 %d", counter)
			}
		})
	}
}
