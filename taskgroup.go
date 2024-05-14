// Package taskgroup 实现了多任务并发执行，聚合收集最终所有任务的执行结果与执行状态，
// 实现了并发（同步）编程中的扇出/扇入模式(Fan-Out/Fan-In)
//
// [taskgroup.TaskGroup] 引用 [sync.WaitGroup], 但增加了任务结果聚合和错误返回（首个必要成功任务的错误信息）
package taskgroup

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"sync"
)

// TaskGroup 表示可将多个任务进行安全并发执行的一个对象
type TaskGroup struct {
	workerNums uint32 // 工作组数量（协程数）

	initOnce       sync.Once
	runExactlyOnce sync.Once // 任务组当且仅当运行一次

	fNOs  map[uint32]struct{}
	tasks []*Task // 待执行的任务集合
}

type Task struct {
	fNO         uint32                      // 任务编号
	f           func() (interface{}, error) // 任务方法
	mustSuccess bool                        // 任务必须执行成功，否则整个任务组将会立即结束，且失败(将会返回第一个必须成功任务的失败结果)
}

// NewTask 创建一个任务
func NewTask(fNO uint32 /* 任务唯一标识 */, f func() (interface{}, error) /* 任务执行方法 */, mustSuccess bool /* 标识任务是否必须执行成功 */) *Task {
	return &Task{fNO, f, mustSuccess}
}

// SetWorkerNums 设置任务所需的协程数
//
// Worker数量的选择：
//
// 1. 硬件资源：系统上的 CPU 核心数量、内存大小和网络带宽等因素会限制可以并行运行的 worker 的数量。如果 worker 数量超过硬件资源能够支持的程度，那么增加更多的 worker 并不会提高整体性能，反而可能因为上下文切换和资源争用而降低性能
//
// 2. 任务的性质：任务可能是 I/O 密集型（如网络请求或磁盘读写）或 CPU 密集型（如复杂的数学计算）。对于 I/O 密集型任务，增加 worker 数量可以更有效地利用等待时间，因为当一个 worker 在等待 I/O 操作完成时，其他 worker 可以继续执行。然而，对于 CPU 密集型任务，增加 worker 数量可能会导致 CPU 资源耗尽，从而降低性能。
//
// 3. 任务的粒度：任务的粒度（即每个任务所需的时间）也会影响 worker 的数量。如果任务粒度很小，那么可能需要更多的 worker 来确保 CPU 始终保持忙碌状态。但是，如果任务粒度很大，那么少量的 worker 就足以处理所有任务，增加 worker 数量可能是不必要的。
func (tg *TaskGroup) SetWorkerNums(workerNums uint32) *TaskGroup {
	tg.workerNums = workerNums
	if tg.workerNums == 0 {
		tg.workerNums = uint32(runtime.NumCPU())
	}
	return tg
}

// AddTask 向任务组中添加若干待执行的任务
func (tg *TaskGroup) AddTask(tasks ...*Task) *TaskGroup {
	tg.initOnce.Do(func() {
		var preAllocatedCapacity = 2*len(tasks) + 1
		tg.fNOs = make(map[uint32]struct{}, preAllocatedCapacity)
		tg.tasks = make([]*Task, 0, preAllocatedCapacity)
	})

	for _, task := range tasks {
		if task == nil {
			continue
		}
		if _, exist := tg.fNOs[task.fNO]; exist {
			panic(fmt.Sprintf("AddTask: Already have the same task %d", task.fNO)) // 已经有相同的任务了
		}
		if task.f != nil {
			tg.fNOs[task.fNO] = struct{}{}
			tg.tasks = append(tg.tasks, task)
		}
	}
	return tg
}

// RunExactlyOnce 启动并运行任务组中的所有任务(仅会运行当且仅当一次)
func (tg *TaskGroup) RunExactlyOnce() (result map[uint32]*taskResult, err error) {
	tg.runExactlyOnce.Do(func() {
		result, err = tg.Run()
	})
	return
}

// Run 启动并运行任务组中的所有任务
func (tg *TaskGroup) Run() (map[uint32]*taskResult, error) {
	taskNums := len(tg.tasks)
	if taskNums == 0 {
		return nil, nil
	}

	// 优先执行必要成功的任务，当同一个goroutine执行多个任务时，如出现了必要成功任务失败时，可提前结束goroutine，即，无需后续任务执行了
	tg.rearrangeTasks()

	var (
		tasks   = make(chan *Task, taskNums)
		results = make(chan *taskResult, taskNums)

		wg   sync.WaitGroup
		once sync.Once
	)
	// 工作协程数不得多余待执行任务总数，否则，因多余协程不会做任务，反而会由于创建或销毁这些协程而带来额外不必要的性能消耗
	if tg.workerNums > uint32(taskNums) {
		tg.workerNums = uint32(taskNums)
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil) // 遍历`ctx`相关的资源泄露(channel, goroutine等)
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

	// fmt.Printf("goroutine nums:%d\n", runtime.NumGoroutine())

	// 发送任务到所有的`workers`中
	for i := 0; i < taskNums; i++ {
		tasks <- tg.tasks[i]
	}
	close(tasks)

	go func() {
		wg.Wait()
		// 当所有任务执行完毕后，再关闭`results`便于能结束遍历收集结果的`for`操作
		close(results)
	}()

	// 从所有的`workers`中收集结果
	taskResults := make(map[uint32]*taskResult, taskNums)
	for result := range results {
		taskResults[result.fNO] = result
	}
	return taskResults, context.Cause(ctx)
}

// rearrangeTasks 任务顺序重排
func (tg *TaskGroup) rearrangeTasks() {
	sort.Slice(tg.tasks, func(i, j int) bool {
		return tg.tasks[i].mustSuccess && !tg.tasks[j].mustSuccess
	})
}

func (tg *TaskGroup) worker(ctx context.Context, tasks chan *Task, results chan *taskResult) error {
	for task := range tasks {
		select {
		case <-ctx.Done(): // 接收到`ctx`被取消的信号，即刻停止后续任务的执行
			return context.Cause(ctx)
		default:
			result, err := task.f() // 每个任务逐一被执行
			if task.mustSuccess && err != nil {
				return err
			}
			results <- &taskResult{task.fNO, result, err}
		}
	}
	return nil
}

type taskResult struct {
	fNO    uint32
	result interface{}
	err    error
}

// Result 获取任务执行结果
func (tr *taskResult) Result() interface{} {
	return tr.result
}

// Error 获取任务执行状态
func (tr *taskResult) Error() error {
	return tr.err
}