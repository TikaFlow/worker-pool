// Package pool 是一个高效、安全、易用的 worker pool 实现
//
// 支持两种使用方式：
//
//  1. 使用默认 pool（推荐，开箱即用）
//     pool.Add(func() { ... })
//     pool.CloseAndWait()
//
//  2. 创建自定义 pool（需要自定义 worker 数量时）
//     myPool := pool.New(20)
//     myPool.Add(func() { ... })
//     myPool.CloseAndWait()
//
// 特性：
//   - 并发安全：无锁设计，高性能
//   - Panic 保护：任务 panic 不会导致 worker 退出，可配置钩子处理
//   - 优雅关闭：支持 Close()（不等待）和 CloseAndWait()（等待所有任务完成）
//   - 默认配置：默认 8 个 worker，16 缓冲通道
//
// 示例：
//
//	// 简单使用
//	pool.Add(func() {
//	    fmt.Println("执行任务")
//	})
//	pool.CloseAndWait()
//
//	// 配置 panic 处理
//	pool.SetPanicHandler(func(r any) {
//	    log.Printf("task panic: %v", r)
//	})
package pool

import (
	"sync"
	"sync/atomic"
)

var (
	defaultWorkerCount int
	defaultPool        *workerPool
)

// PanicHandler 定义 panic 处理函数类型
type PanicHandler func(any)

func init() {
	defaultWorkerCount = 8
	defaultPool = New(defaultWorkerCount)
}

// workerTask 无参无返回值的任务函数
type workerTask func()

// workerPool 保存worker pool的数据
type workerPool struct {
	taskCh       chan workerTask // 任务通道
	workerCount  int             // 并发上限
	closeOnce    sync.Once       // 保证 Close 只执行一次
	wg           sync.WaitGroup  // 等待所有 worker 退出
	panicHandler atomic.Value    // panic 处理钩子，原子操作保证安全
}

// New 创建一个新的worker pool
// workerCount: 并发上限，必须大于0
func New(workerCount int) *workerPool {
	// 检查参数合法性
	if workerCount <= 0 {
		workerCount = 1
	}

	pool := &workerPool{
		taskCh:      make(chan workerTask, workerCount*2),
		workerCount: workerCount,
	}

	// 启动worker
	pool.run()

	return pool
}

// SetPanicHandler 设置默认 worker pool 的 panic 处理函数
func SetPanicHandler(handler PanicHandler) {
	defaultPool.SetPanicHandler(handler)
}

// SetPanicHandler 设置 panic 处理函数
func (wp *workerPool) SetPanicHandler(handler PanicHandler) {
	wp.panicHandler.Store(handler)
}

// Add 向默认worker pool添加任务
func Add(task workerTask) {
	if task == nil {
		return
	}
	defaultPool.Add(task)
}

// Add 向worker pool添加任务
func (wp *workerPool) Add(task workerTask) {
	if task == nil {
		return
	}

	func() {
		defer func() {
			recover() // 静默 recover 向已关闭通道发送的 panic
		}()
		wp.taskCh <- task
	}()
}

// Close 关闭默认worker pool
func Close() error {
	return defaultPool.Close()
}

// Close 关闭worker pool，实现io.Closer接口
func (wp *workerPool) Close() error {
	wp.closeOnce.Do(func() {
		close(wp.taskCh)
	})
	return nil
}

// CloseAndWait 关闭默认worker pool并等待所有worker退出
func CloseAndWait() error {
	return defaultPool.CloseAndWait()
}

// CloseAndWait 关闭worker pool并等待所有worker退出
func (wp *workerPool) CloseAndWait() error {
	wp.Close()
	wp.wg.Wait()
	return nil
}

// worker 私有方法：worker执行任务
func (wp *workerPool) worker() {
	defer wp.wg.Done()
	for task := range wp.taskCh {
		func() {
			defer func() {
				if r := recover(); r != nil {
					if handler, ok := wp.panicHandler.Load().(PanicHandler); ok && handler != nil {
						handler(r)
					}
				}
			}()
			task()
		}()
	}
}

// run 私有函数：启动worker，启动个数为并发上限
func (wp *workerPool) run() {
	wp.wg.Add(wp.workerCount)
	for i := 0; i < wp.workerCount; i++ {
		go wp.worker()
	}
}
