# taskgroup

*golang*并发执行多任务，并聚合多任务结果。
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
  > > 给[errgroup](https://cs.opensource.google/go/x/sync)项目提PR
- **errgroup**可支持带有[取消Context](https://pkg.go.dev/context#WithCancelCause)的模式，但实际上，该种模式下仍需要所有执行任务的`goroutine`执行完毕（每一个任务都会有新的`goroutine`）

# IDEAs & TODOs

1. 考虑所有操作支持并发安全, 比如，`TaskGroup.fNOs`, `TaskGroup.tasks`
2. 考虑下关键的结构对象使用指针还是非指针结构(在结构体对象的大小和结构体对象的总体数量上做下权衡，如果产生结构体对象会较多，使用指针堆对象可能会带来`GC`压力，如果结构体对象本身复杂，申请栈对象可能带来较大的额外内存复制的开销)
3. 补充`benchmark`测试，以及和`errgroup`性能对比
4. 进行详细的性能分析，`go tool trace; go tool pprof` 完整的分析协程调度细节、cpu、内存使用情况（火焰图）
5. 补充`Example`，严格遵循 [godoc](https://pkg.go.dev/) 文档&注释规范
