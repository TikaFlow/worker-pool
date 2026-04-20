// Package pool 是一个高效、安全、易用的 worker pool 实现
//
// 使用方式：
//
//	myPool := pool.New(20, nil)
//	myPool.Add(func() { ... })
//	myPool.CloseAndWait()
//
// 特性：
//   - 并发安全：无锁设计，高性能
//   - Panic 保护：任务 panic 不会导致 worker 退出，可配置钩子处理
//   - 优雅关闭：支持 Close()（不等待）和 CloseAndWait()（等待所有任务完成）
//
// 示例：
//
//	// 简单使用
//	p := pool.New(8, nil)
//	p.Add(func() {
//	    fmt.Println("执行任务")
//	})
//	p.CloseAndWait()
//
//	// 配置 panic 处理
//	p := pool.New(8, &pool.Config{
//	    PanicHandler: func(r any) {
//	        log.Printf("task panic: %v", r)
//	    },
//	})
package pool

import (
	"sync"
)

// Config 配置项
type Config struct {
	// PanicHandler 任务 panic 处理函数
	PanicHandler func(any)
}

// workerTask 无参无返回值的任务函数
type workerTask func()

// workerPool 保存 worker pool 的数据
type workerPool struct {
	taskCh       chan workerTask
	workerCount  int
	closeOnce    sync.Once
	wg           sync.WaitGroup
	panicHandler func(any)
}

// New 创建一个新的 worker pool
// workerCount: 并发上限，必须大于 0
// cfg: 配置项，nil 表示无额外配置
func New(workerCount int, cfg *Config) *workerPool {
	if workerCount <= 0 {
		workerCount = 1
	}

	p := &workerPool{
		taskCh:       make(chan workerTask, workerCount*2),
		workerCount:  workerCount,
		panicHandler: func(any) {}, // 默认空函数，确保不为 nil
	}

	if cfg != nil && cfg.PanicHandler != nil {
		p.panicHandler = cfg.PanicHandler
	}

	p.start()
	return p
}

// Add 向 worker pool 添加任务
func (wp *workerPool) Add(task workerTask) {
	if task == nil {
		return
	}

	func() {
		defer func() {
			recover()
		}()
		wp.taskCh <- task
	}()
}

// Close 关闭 worker pool
func (wp *workerPool) Close() error {
	wp.closeOnce.Do(func() {
		close(wp.taskCh)
	})
	return nil
}

// CloseAndWait 关闭 worker pool 并等待所有 worker 退出
func (wp *workerPool) CloseAndWait() error {
	wp.Close()
	wp.wg.Wait()
	return nil
}

// worker 私有方法：worker 执行任务
func (wp *workerPool) worker() {
	defer wp.wg.Done()
	for task := range wp.taskCh {
		func() {
			defer func() {
				if r := recover(); r != nil {
					wp.panicHandler(r)
				}
			}()
			task()
		}()
	}
}

// start 私有方法：启动 worker
func (wp *workerPool) start() {
	wp.wg.Add(wp.workerCount)
	for i := 0; i < wp.workerCount; i++ {
		go wp.worker()
	}
}
