# taskgroup
[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue "godoc")](https://pkg.go.dev/github.com/mlee-msl/taskgroup)  

*golang*并发执行多任务(网络i/o、磁盘i/o、内存i/o等)，并聚合收集多任务执行结果与执行状态（如<u>任务组执行失败</u>，将返回首个<u>必要成功</u>任务的错误信息, 且会立即停止后续任务的运行）。
**[`Quick Start`](https://pkg.go.dev/github.com/mlee-msl/taskgroup "欢迎使用，任何意见或建议可联系`2210508401@qq.com`")** & **[`Best Practice`](https://github.com/mlee-msl/taskgroup/blob/d92c0b0ca072c94be86e7147338d77d3707fd0b5/taskgroup_example_test.go#L15)**

> **使用:** go get github.com/mlee-msl/taskgroup

## 功能特点

- 并发安全的执行多个任务
- 将多个任务的执行结果与执行状态进行聚合
- 通过**扇出/扇入**模式，结合线程安全`channel`实现高效协程间通信
- <big>多任务复用(共享)同一协程</big>，避免了因协程频繁创建或销毁带来的开销，一定程度上也减少了上下文切换（特别是，内核线程与用户协程间的切换）的频次
- `early-return`，当出现<big><u>必要成功</u></big>的任务失败时，将停止执行所有`goroutine`上还未启动的所有其他任务
  >NOTEs，当所有任务都设置为非必要成功时，即可退化为`errgroup`包的使用场景

## vs官方扩展库[errgroup](https://pkg.go.dev/golang.org/x/sync/errgroup "errgroup")

- **errgroup** 没有任务添加阶段，直接会使用协程执行指定的任务
  > 可通过限制协程数量上限，控制并发量（指定`buffer size`的`channel`实现）,当协程数达到上限时，需要等待现有任务执行结束，然后开启新的协程，会增加协程创建或销毁的成本
  >
  >> 另外，当协程量不指定时（默认协程量不受限），随着任务量的增加，会增加内核态线程与用户态协程间的上下文切换成本
- **errgroup**可支持带有[取消Context](https://pkg.go.dev/context#WithCancelCause)的模式，但实际上，该种模式下仍需要所有执行任务的`goroutine`执行完毕（每个任务都会绑定到新的`goroutine`上）
- 仅返回了首个出现错误的接口(任务)状态，未将所有接口(任务)的执行结果（或执行状态）进行收集

## IDEAs & TODOs

- 当使用[errgroup.WithContext](https://cs.opensource.google/go/x/sync/+/master:errgroup/errgroup.go;l=48;bpv=1;bpt=1)时，出现错误`cancel`后，需要同步（如，`atomic`同步原语）后续协程<big><u>（因[任务阻塞](https://cs.opensource.google/go/x/sync/+/master:errgroup/errgroup.go;l=71;bpv=1;bpt=1)还未开始执行的协程）</u></big>不再启动(后续执行已无意义，<u>业务层面已选择了带取消的上下文执行方式</u>), 也确保避免了出现内存泄露（协程泄露等）等可能的问题
- 增加一个默认的`context.Context`，和`errgroup.WithContext`的处理逻辑保持一致，统一使用`context.Context`处理错误 (功能实现上，整体可能会更加优雅点）

## 关于性能
### 指标

> `cpu`性能  
>> `go test -benchmem -run=^$ -bench ^BenchmarkTaskGroup$ -cpu='1,2,4,8,16' [-benchtime=10x|-benchtime=1s] -count=3 -cpuprofile='cpu.pprof' .`   
>
> 内存性能  
>> `go test -benchmem -run=^$ -bench ^BenchmarkTaskGroup$ -cpu='1,2,4,8,16' [-benchtime=10x|-benchtime=1s] -count=3 -memprofile='mem.pprof' .`  
>
> 阻塞性能  
>> `go test -benchmem -run=^$ -bench ^BenchmarkTaskGroup$ -cpu='1,2,4,8,16' [-benchtime=10x|-benchtime=1s] -count=3 -blockprofile='block.pprof' .`   
>
> 锁性能  
>>`go test -benchmem -run=^$ -bench ^BenchmarkTaskGroup$ -cpu='1,2,4,8,16' [-benchtime=10x|-benchtime=1s] -count=3 -mutexprofile='mutex.pprof' .`   
### 效果对比
> `taskgroup`   
>> 并发任务量: 6 
>> ![taskgroup_6.jpg](https://s2.loli.net/2024/05/23/rnBPEpqkZCNJYvz.jpg)   
>> 并发任务量: 60  
>> ![taskgroup_60.jpg](https://s2.loli.net/2024/05/23/oviDfZxEnpyCBsc.jpg)   
>
> `errgroup`   
>> 并发任务量: 6   
>> ![errgroup_6.jpg](https://s2.loli.net/2024/05/23/bYDLAC7nMx6WKPv.jpg)   
>> 并发任务量: 60   
>> ![errgroup_60.jpg](https://s2.loli.net/2024/05/23/mJKscFZ5pyhHal7.jpg)   
>
>> <big>**可以看出**</big>，在增加任务执行结果聚合、失败快速返回的能力前提下：  
>>> 当并发任务量不多时，前者相较于后者的综合性能提升还不明显；     
>>> 当并发任务量较多时，前者相较于后者的综合性能表现提升较大；   
>>> 深度验证后发现，随着并发任务量的增加，其综合性能表现的差距将会越明显。   
## 其他
- 核心功能的测试用例均采用`Fuzz Test`模糊测试，并测试通过
- 包中提供的核心功能，均有`Example Test`样例用例，并执行通过，**[`更多详情`](https://pkg.go.dev/github.com/mlee-msl/taskgroup "欢迎使用，任何意见或建议可联系`2210508401@qq.com`")**
