# Worker Pool

一个高效、安全、易用的 Go 语言 Worker Pool 实现。

## 特性

- **并发安全**：无锁设计，高性能
- **Panic 保护**：任务 panic 不会导致 worker 退出，可配置钩子处理
- **优雅关闭**：支持 Close()（不等待）和 CloseAndWait()（等待所有任务完成）
- **默认配置**：默认 8 个 worker，16 缓冲通道

## 安装

```bash
go get github.com/TikaFlow/worker-pool
```

## 快速开始

### 使用默认 Pool（推荐）

```go
package main

import (
    "fmt"
    "log"

    "github.com/TikaFlow/worker-pool"
)

func main() {
    // 添加任务
    pool.Add(func() {
        fmt.Println("执行任务 1")
    })

    pool.Add(func() {
        fmt.Println("执行任务 2")
    })

    // 配置 panic 处理（可选）
    pool.SetPanicHandler(func(r any) {
        log.Printf("任务 panic: %v", r)
    })

    // 关闭并等待所有任务完成
    pool.CloseAndWait()
}
```

### 使用自定义 Pool

```go
package main

import (
    "fmt"

    "github.com/TikaFlow/worker-pool"
)

func main() {
    // 创建自定义 pool（20 个 worker）
    myPool := pool.New(20)

    // 添加任务
    myPool.Add(func() {
        fmt.Println("执行任务")
    })

    // 关闭并等待
    myPool.CloseAndWait()
}
```

## API 文档

### 默认 Pool

| 函数 | 说明 |
|------|------|
| `Add(task func())` | 添加任务 |
| `SetPanicHandler(handler func(any))` | 设置 panic 处理函数 |
| `Close() error` | 关闭（不等待） |
| `CloseAndWait() error` | 关闭并等待所有任务完成 |

### 自定义 Pool

| 函数 | 说明 |
|------|------|
| `New(workerCount int) *workerPool` | 创建自定义 pool |
| `(wp *workerPool) Add(task func())` | 添加任务 |
| `(wp *workerPool) SetPanicHandler(handler func(any))` | 设置 panic 处理函数 |
| `(wp *workerPool) Close() error` | 关闭（不等待） |
| `(wp *workerPool) CloseAndWait() error` | 关闭并等待所有任务完成 |

## 设计特点

1. **懒加载关闭保护**：Close() 后调用 Add() 会被静默忽略，不会 panic
2. **并发安全**：使用 atomic.Value 和 sync.Once，无数据竞争
3. **高性能**：无锁设计，直接通过 channel 通信
4. **默认自动初始化**：包导入时自动创建默认 pool，开箱即用

## 许可证

MIT License

Copyright (c) 2026 兮夏

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
