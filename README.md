# taskgroup

*golang*并发执行多任务，并聚合收集多任务执行结果与执行状态（如<u>任务组执行失败</u>，返回首个必要成功任务的错误信息, 且会立即停止后续任务的运行）。
**[`使用文档`](https://pkg.go.dev/github.com/mlee-msl/taskgroup "欢迎使用，任何意见或建议可联系`2210508401@qq.com`")**

> **使用:** go get github.com/mlee-msl/taskgroup

# 功能特点

- 并发安全的执行多个任务
- 将多个任务的结果进行聚合
- 通过**扇出/扇入**模式，结合线程安全`channel`实现高效协程间通信
- <big>多任务复用(共享)同一协程</big>，避免了协程频繁创建或销毁的开销
- `early-return`，当出现<big><u>必要成功</u></big>的任务失败时，停止执行所有`goroutine`后续的其他任务

# 对比[errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup "errgroup")

- **errgroup** 没有任务添加阶段，直接会使用协程执行指定的任务
  > 可通过限制协程数量上限，控制并发量（指定`buffer size`的`channel`实现）,当协程数达到上限时，需要等待现有任务执行结束，然后开启新的协程，会增加协程创建或销毁的成本
  >
  >> 另外，当协程量不指定时（默认协程量不受限），随着任务量的增加，会增加内核态线程与用户态协程间的上下文切换成本
- **errgroup**可支持带有[取消Context](https://pkg.go.dev/context#WithCancelCause)的模式，但实际上，该种模式下仍需要所有执行任务的`goroutine`执行完毕（每一个任务都会有新的`goroutine`）

# IDEAs & TODOs

1. 提[errgroup](https://cs.opensource.google/go/x/sync)项目Issue

# 关于性能

- `cpu`性能  
`go test -benchmem -run=^$ -bench ^BenchmarkTaskGroupHigh$ -cpu='1,2,4,8,16' -benchtime=10x -count=5 -cpuprofile='cpu.pprof' .`
- 内存性能  
`go test -benchmem -run=^$ -bench ^BenchmarkTaskGroupHigh$ -cpu='1,2,4,8,16' -benchtime=10x -count=5 -memprofile='mem.pprof' .`
- 阻塞性能  
`go test -benchmem -run=^$ -bench ^BenchmarkTaskGroupHigh$ -cpu='1,2,4,8,16' -benchtime=10x -count=5 -blockprofile='block.pprof' .`
- 锁性能  
`go test -benchmem -run=^$ -bench ^BenchmarkTaskGroupHigh$ -cpu='1,2,4,8,16' -benchtime=10x -count=5 -mutexprofile='mutex.pprof' .`