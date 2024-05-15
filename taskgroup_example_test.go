package taskgroup_test

import (
	"fmt"

	"github.com/mlee-msl/taskgroup"
)

// ---------------------------------------------------
// ----------------------使用样例----------------------
// ---------------------------------------------------
// Default 展示了默认配置的使用案例，包括，多任务创建、任务执行、结果收集，错误处理等
func ExampleTaskGroup_default() {
	var (
		tg *taskgroup.TaskGroup

		tasks = []*taskgroup.Task{
			taskgroup.NewTask(1, task1ReturnFailWrapper(1), true),
			taskgroup.NewTask(2, task2ReturnSuccessWrapper(2), false),
			taskgroup.NewTask(3, task3ReturnFailWrapper(3), true),
		}
	)
	tg = taskgroup.NewTaskGroup()
	taskResults, err := tg.AddTask(tasks...).Run()
	if err != nil {
		fmt.Printf("err: %+v", err)
		return
	}
	for fno, result := range taskResults {
		fmt.Printf("FNO: %d, RESULT: %v , STATUS: %v\n", fno, result.Result(), result.Error())
	}
}

// JustErrors 展示了错误（非最佳）的使用案例，包括，多任务创建、任务执行、结果收集，错误处理等
func ExampleTaskGroup_justErrors() {
	var (
		tg *taskgroup.TaskGroup

		tasks = []*taskgroup.Task{
			taskgroup.NewTask(1, task1ReturnFailWrapper(1), true),
			taskgroup.NewTask(2, task2ReturnSuccessWrapper(2), false),
			taskgroup.NewTask(3, task3ReturnFailWrapper(3), true),
			nil,
			nil,
		}
	)
	tg = taskgroup.NewTaskGroup(nil)
	taskResults, err := tg.AddTask(tasks...).Run()
	if err != nil {
		fmt.Printf("err: %+v", err)
		return
	}
	for fno, result := range taskResults {
		fmt.Printf("FNO: %d, RESULT: %v , STATUS: %v\n", fno, result.Result(), result.Error())
	}
}

// JustAbnormal 展示了异常的使用案例，包括，多任务创建、任务执行、结果收集，错误处理等
func ExampleTaskGroup_justAbnormal() {
	var (
		tg *taskgroup.TaskGroup

		tasks = []*taskgroup.Task{
			taskgroup.NewTask(1, task1ReturnFailWrapper(1), true),
			taskgroup.NewTask(2, task2ReturnSuccessWrapper(2), false),
			taskgroup.NewTask(3, task3ReturnFailWrapper(3), true),
			nil,
			nil,
		}
	)
	taskResults, err := tg.AddTask(tasks...).Run()
	if err != nil {
		fmt.Printf("err: %+v", err)
		return
	}
	for fno, result := range taskResults {
		fmt.Printf("FNO: %d, RESULT: %v , STATUS: %v\n", fno, result.Result(), result.Error())
	}
}

// Typical 展示了典型的使用案例，包括，多任务创建、任务执行、结果收集，错误处理等
func ExampleTaskGroup_typical() {
	var (
		tg *taskgroup.TaskGroup

		tasks = []*taskgroup.Task{
			taskgroup.NewTask(1, task1ReturnFailWrapper(1), true),
			taskgroup.NewTask(2, task2ReturnSuccessWrapper(2), false),
			taskgroup.NewTask(3, task3ReturnFailWrapper(3), false),
		}
	)
	tg = taskgroup.NewTaskGroup(taskgroup.WithWorkerNums(6))
	taskResults, err := tg.AddTask(tasks...).Run()
	if err != nil {
		fmt.Printf("err: %+v", err)
		return
	}
	for fno, result := range taskResults {
		fmt.Printf("FNO: %d, RESULT: %v , STATUS: %v\n", fno, result.Result(), result.Error())
	}
}
