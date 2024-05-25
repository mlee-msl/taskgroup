// Package taskgroup 实现了多任务并发执行，聚合收集最终所有任务的执行结果与执行状态，
//
// [taskgroup.TaskGroup] 引用 [sync.WaitGroup], 但增加了任务结果聚合和错误返回（首个必要成功任务的错误信息, 且会立即停止后续任务的运行）
package taskgroup

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// TaskGroup 表示可将多个任务进行安全并发执行的一个对象
type TaskGroup struct {
	workerNums uint32 // 工作组数量（协程数）

	initOnce       sync.Once
	runExactlyOnce sync.Once

	fNOs  map[uint32]struct{}
	tasks []*Task
}

// TaskFunc 任务函数的签名
type TaskFunc func() (interface{}, error)

type Task struct {
	fNO         uint32   // 任务编号(标识)
	f           TaskFunc // 任务方法
	mustSuccess bool     // 任务必须执行成功，否则整个任务组将会立即结束，且失败(将会返回第一个必须成功任务的失败结果)
}

// Option 表示任务组对象默认行为的修改
type Option func(*TaskGroup)

// WithWorkerNums 指定任务组执行时所需的协程数`workerNums`
//
// Worker 数量的选择：
//
// 1. 硬件资源：系统上的 CPU 核心数量、内存大小和网络带宽等因素会限制可以并行运行的 worker 的数量。如果 worker 数量超过硬件资源能够支持的程度，那么增加更多的 worker 并不会提高整体性能，反而可能因为上下文切换和资源争用而降低性能
//
// 2. 任务的性质：任务可能是 I/O 密集型（如网络请求或磁盘读写）或 CPU 密集型（如复杂的数学计算）。对于 I/O 密集型任务，增加 worker 数量可以更有效地利用等待时间，因为当一个 worker 在等待 I/O 操作完成时，其他 worker 可以继续执行。然而，对于 CPU 密集型任务，增加 worker 数量可能会导致 CPU 资源耗尽，从而降低性能。
//
// 3. 任务的粒度：任务的粒度（即每个任务所需的时间）也会影响 worker 的数量。如果任务粒度很小，那么可能需要更多的 worker 来确保 CPU 始终保持忙碌状态。但是，如果任务粒度很大，那么少量的 worker 就足以处理所有任务，增加 worker 数量可能是不必要的。
func WithWorkerNums(workerNums uint32) Option {
	return func(tg *TaskGroup) {
		if tg == nil {
			return
		}
		tg.workerNums = workerNums
	}
}

// NewTaskGroup 创建一个任务组对象
func NewTaskGroup(opts ...Option) *TaskGroup {
	tg := new(TaskGroup)
	for _, opt := range opts {
		if opt != nil {
			opt(tg)
		}
	}
	taskNums := (tg.workerNums + 1) * 2 // 预留协程数2倍之有余的任务数
	tg.fNOs = make(map[uint32]struct{}, taskNums)
	tg.tasks = make([]*Task, 0, taskNums)
	return tg
}

// NewTask 创建一个任务，`fNO`用以表示任务`f`的唯一标识, `mustSuccess`则表示该任务`f`是否为必须成功，当`true`时,
// 且任务`f`执行失败，表示整个任务组将执行失败
func NewTask(fNO uint32 /* 任务唯一标识 */, f TaskFunc /* 任务执行方法 */, mustSuccess bool /* 标识任务是否必须执行成功 */) *Task {
	return &Task{fNO, f, mustSuccess}
}

// AddTask 向任务组中添加若干待执行的任务`tasks`
//
// NOTEs: 出现了相同的任务(任务的标识相等)，将会`panic`
func (tg *TaskGroup) AddTask(tasks ...*Task) *TaskGroup {
	if tg == nil {
		return nil
	}

	if tg.fNOs == nil || len(tasks) > cap(tg.tasks) {
		tg.initOnce.Do(func() {
			preAllocatedCapacity := (len(tasks) + 1) * 2
			tg.fNOs = make(map[uint32]struct{}, preAllocatedCapacity)
			tg.tasks = make([]*Task, 0, preAllocatedCapacity)
		})
	}

	for i := 0; i < len(tasks); i++ {
		if tasks[i] == nil || tasks[i].f == nil {
			continue
		}

		if _, exist := tg.fNOs[tasks[i].fNO]; exist { // 已经有相同的任务了
			panic(fmt.Sprintf("AddTask: Already have the same Task %d", tasks[i].fNO))
		}

		tg.fNOs[tasks[i].fNO] = struct{}{}
		tg.tasks = append(tg.tasks, tasks[i])
	}
	return tg
}

// RunExactlyOnce 启动并运行任务组中的所有任务(运行当且仅当一次)
func (tg *TaskGroup) RunExactlyOnce() (result map[uint32]*TaskResult, err error) {
	if tg == nil {
		return nil, nil
	}

	tg.runExactlyOnce.Do(func() {
		result, err = tg.Run()
	})
	return
}

// Run 启动并运行任务组中的所有任务
//
// 当返回`non-nil`错误时，则，返回的任务执行结果将不可信
func (tg *TaskGroup) Run() (map[uint32]*TaskResult, error) {
	if tg == nil {
		return nil, nil
	}

	taskNums := len(tg.tasks)
	if taskNums == 0 {
		return nil, nil
	}

	// 执行任务前的若干准备工作
	tg.prepare()

	var (
		tasks   = make(chan *Task, taskNums)
		results = make(chan *TaskResult, taskNums)

		wg   sync.WaitGroup
		once sync.Once
	)
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil) // 避免`ctx`相关的资源泄露(channel, goroutine等)
	// 启动`workers`
	for i := 1; i <= int(tg.workerNums); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := tg.worker(ctx, tasks, results); err != nil {
				once.Do(func() {
					cancel(err)
				})
			}
		}()
	}

	// 发送任务到所有的`workers`中
	for i := 0; i < taskNums; i++ {
		tasks <- tg.tasks[i]
	}
	close(tasks)

	go func() {
		wg.Wait()
		close(results)
	}()

	// 从所有的`workers`中收集结果
	taskResults := make(map[uint32]*TaskResult, taskNums)
	for result := range results {
		taskResults[result.fNO] = result
	}
	return taskResults, context.Cause(ctx)
}

func (tg *TaskGroup) prepare() {
	// 优先执行必要成功的任务，当同一个goroutine执行多个任务时，如出现了必要成功任务失败时，可提前结束goroutine，即，无需后续任务执行了
	rearrangeTasks(tg.tasks)
	// 调整工作组中的协程量
	WithWorkerNums(adjustWorkerNums(tg.workerNums, uint32(len(tg.tasks))))(tg)
}

// rearrangeTasks 任务顺序重排
//
//go:nosplit
func rearrangeTasks(tasks []*Task) {
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i] != nil && tasks[j] != nil {
			return tasks[i].mustSuccess && !tasks[j].mustSuccess
		}
		return true
	})
}

// adjustWorkerNums 调整工作组中的协程量
//
//go:nosplit
func adjustWorkerNums(workerNums, taskNums uint32) uint32 {
	// 工作协程数不得多余待执行任务总数，否则，因多余协程不会做任务，反而会由于创建或销毁这些协程而带来额外不必要的性能消耗
	workerNums = If(workerNums > taskNums, taskNums, workerNums).(uint32)
	const taskShareLimit = 5 // 任务共享协程量上限
	if minWorkerNums := (taskNums / taskShareLimit) + 1; workerNums < minWorkerNums {
		// 当任务数过多，协程数又过少时，会出现任务执行等待(by Fuzz Test)
		workerNums = minWorkerNums
	}
	return workerNums
}

// worker 若干个任务将会共享在一个协程上执行任务
func (tg *TaskGroup) worker(ctx context.Context, tasks <-chan *Task, results chan<- *TaskResult) error {
	for task := range tasks {
		select {
		case <-ctx.Done(): // 接收到`ctx`被取消的信号，即刻停止后续任务的执行
			return context.Cause(ctx)
		default:
			result, err := task.f()
			if task.mustSuccess && err != nil {
				return err
			}
			// 防止向关闭的`channel`中写入数据
			if context.Cause(ctx) == nil {
				results <- &TaskResult{task.fNO, result, err}
			}
		}
	}
	return nil
}

// TaskResult 表示任务的执行结果与执行状态
type TaskResult struct {
	fNO    uint32
	result interface{}
	err    error
}

// FNO 获取任务的唯一标识号
func (tr *TaskResult) FNO() uint32 {
	if tr == nil {
		return 0
	}
	return tr.fNO
}

// Result 获取任务执行结果
func (tr *TaskResult) Result() interface{} {
	if tr == nil {
		return nil
	}
	return tr.result
}

// Error 获取任务执行状态
func (tr *TaskResult) Error() error {
	if tr == nil {
		return nil
	}
	return tr.err
}

// If 简单的三元表达式实现
var If = func(cond bool, a, b interface{}) interface{} {
	if cond {
		return a
	}
	return b
}
