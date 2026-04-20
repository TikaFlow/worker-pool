// Package pool 是一个高效、安全、易用的 worker pool 实现
//
// 使用方式：
//
//	myPool := pool.New(20, nil)
//	myPool.Add(func() { ... })
//	myPool.Close()
//
// 特性：
//   - 并发安全：无锁设计，高性能
//   - Panic 保护：任务 panic 不会导致 worker 退出，可配置钩子处理
//   - 优雅关闭：支持 Close()（等待）和 CloseNoWait()（不等待）
//
// 示例：
//
//	// 简单使用
//	p := pool.New(8, nil)
//	p.Add(func() {
//	    fmt.Println("执行任务")
//	})
//	p.Close()
//
//	// 配置 panic 处理
//	p := pool.New(8, &pool.Config{
//	    PanicHandler: func(r any) {
//	        log.Printf("task panic: %v", r)
//	    },
//	})
package pool

import (
	"io"
	"sync"
)

type Pool interface {
	io.Closer
	Add(task workerTask)
	CloseNoWait() error
}

type Config struct {
	PanicHandler func(any)
	BufferSize   int
}

type workerTask func()

type workerPool struct {
	taskCh       chan workerTask
	workerCount  int
	closeOnce    sync.Once
	wg           sync.WaitGroup
	panicHandler func(any)
}

func New(workerCount int, cfg *Config) Pool {
	if workerCount <= 0 {
		workerCount = 1
	}

	bufferSize := workerCount * 2
	if bufferSize < 16 {
		bufferSize = 16
	}

	if cfg != nil && cfg.BufferSize > 0 {
		bufferSize = cfg.BufferSize
	}

	p := &workerPool{
		taskCh:       make(chan workerTask, bufferSize),
		workerCount:  workerCount,
		panicHandler: func(any) {},
	}

	if cfg != nil && cfg.PanicHandler != nil {
		p.panicHandler = cfg.PanicHandler
	}

	p.start()
	return p
}

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

func (wp *workerPool) Close() error {
	wp.CloseNoWait()
	wp.wg.Wait()
	return nil
}

func (wp *workerPool) CloseNoWait() error {
	wp.closeOnce.Do(func() {
		close(wp.taskCh)
	})
	return nil
}

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

func (wp *workerPool) start() {
	wp.wg.Add(wp.workerCount)
	for i := 0; i < wp.workerCount; i++ {
		go wp.worker()
	}
}
