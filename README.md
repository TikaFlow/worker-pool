# Worker Pool

一个高效、安全、易用的 Go 语言 Worker Pool 实现。

## 特性

- **并发安全**：无锁设计，高性能
- **Panic 保护**：任务 panic 不会导致 worker 退出，可配置钩子处理
- **优雅关闭**：支持 Close()（等待）和 CloseNoWait()（不等待）
- **接口设计**：返回接口类型，方便定义变量

## 安装

```bash
go get github.com/TikaFlow/worker-pool
```

## 快速开始

```go
package main

import (
	"fmt"

	"github.com/TikaFlow/worker-pool"
)

func main() {
	p := pool.New(8, nil)
	defer p.Close()

	p.Add(func() {
		fmt.Println("执行任务 1")
	})

	p.Add(func() {
		fmt.Println("执行任务 2")
	})
}
```

### 使用 Panic Handler

```go
package main

import (
	"log"

	"github.com/TikaFlow/worker-pool"
)

func main() {
	p := pool.New(8, &pool.Config{
		PanicHandler: func(r any) {
			log.Printf("任务 panic: %v", r)
		},
	})
	defer p.Close()

	p.Add(func() {
		panic("测试 panic")
	})
}
```

## API 文档

| 函数/类型 | 说明 |
|----------|------|
| `New(workerCount int, cfg *Config) Pool` | 创建 worker pool，返回接口类型 |
| `Config` | 配置项结构体 |
| `Config.PanicHandler` | Panic 处理函数 |
| `Config.BufferSize` | 任务通道缓冲大小，默认取 `workerCount*2` 与 16 的较大值 |
| `Pool.Add(task func())` | 添加任务 |
| `Pool.Close() error` | 关闭并等待所有 worker 退出，实现 io.Closer |
| `Pool.CloseNoWait() error` | 关闭但不等待 worker 退出 |

## 设计特点

1. **懒加载关闭保护**：Close() 后调用 Add() 会被静默忽略，不会 panic
2. **并发安全**：使用 sync.Once，无数据竞争
3. **高性能**：无锁设计，直接通过 channel 通信
4. **接口设计**：返回接口类型，便于面向接口编程
5. **配置简洁**：结构体配置，nil 表示无额外配置

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
