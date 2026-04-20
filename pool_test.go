package pool

import (
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	p := New(5, nil)
	if p == nil {
		t.Fatal("New 返回 nil")
	}
	p.Close()
}

func TestNewWithZero(t *testing.T) {
	p := New(0, nil)
	if p == nil {
		t.Fatal("New 返回 nil")
	}
	p.Close()
}

func TestNewWithNegative(t *testing.T) {
	p := New(-10, nil)
	if p == nil {
		t.Fatal("New 返回 nil")
	}
	p.Close()
}

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
	p.Close()

	if counter != 100 {
		t.Errorf("期望 100，实际 %d", counter)
	}
}

func TestAddNil(t *testing.T) {
	p := New(4, nil)
	p.Add(nil)
	p.Close()
}

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
	p.Close()

	if panicCaught != "test panic" {
		t.Errorf("期望 'test panic'，实际 %v", panicCaught)
	}
}

func TestCloseMultipleTimes(t *testing.T) {
	p := New(4, nil)
	p.Close()
	p.Close()
	p.Close()
}

func TestClose(t *testing.T) {
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

	p.Close()

	if counter != taskCount {
		t.Errorf("期望 %d，实际 %d", taskCount, counter)
	}
}

func TestAddAfterClose(t *testing.T) {
	p := New(4, nil)
	p.Close()
	time.Sleep(10 * time.Millisecond)
	p.Add(func() {})
}

func TestConcurrent(t *testing.T) {
	p := New(16, nil)
	taskCount := 10000
	var wg sync.WaitGroup

	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		p.Add(func() {
			time.Sleep(50 * time.Microsecond)
			wg.Done()
		})
	}

	wg.Wait()
	p.Close()
}

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
			p.Close()

			if counter != 100 {
				t.Errorf("期望 100，实际 %d", counter)
			}
		})
	}
}
