package taskgroup_test

import (
	"fmt"

	"github.com/mlee-msl/taskgroup"
)

// ---------------------------------------------------
// ----------------------使用样例----------------------
// ---------------------------------------------------
// Typical 展示了典型的使用案例，包括，多任务创建、任务执行、结果收集，错误处理等
func ExampleTaskGroup_typical() {
	tasks := []*taskgroup.Task{
		taskgroup.NewTask(1, task1ReturnFailWrapper(1, false), false),
		taskgroup.NewTask(2, task2ReturnSuccessWrapper(2, false), true),
		taskgroup.NewTask(3, task3ReturnFailWrapper(3, false), false),
	}

	taskResults, err := taskgroup.NewTaskGroup(taskgroup.WithWorkerNums(uint32(len(tasks)))).AddTask(tasks...).Run()
	if err != nil {
		fmt.Printf("err: %+v\n", err)
		return
	}
	for fno, result := range taskResults {
		fmt.Printf("FNO: %d, RESULT: %v , STATUS: %v\n", fno, result.Result(), result.Error())
	}
	// Unordered output:
	// FNO: 3, RESULT: TASK3: The data is 928 , STATUS: fno: 3, TASK3 err
	// FNO: 2, RESULT: {1112 mlee} , STATUS: <nil>
	// FNO: 1, RESULT: 1127 , STATUS: fno: 1, TASK1 err
}

// Default 展示了默认配置的使用案例，包括，多任务创建、任务执行、结果收集，错误处理等
func ExampleTaskGroup_default() {
	tasks := []*taskgroup.Task{
		taskgroup.NewTask(1, task1ReturnFailWrapper(1, false), false),
		taskgroup.NewTask(2, task2ReturnSuccessWrapper(2, false), true),
		taskgroup.NewTask(3, task3ReturnFailWrapper(3, false), true),
	}

	taskResults, err := new(taskgroup.TaskGroup).AddTask(tasks...).Run()
	if err != nil {
		fmt.Printf("err: %+v\n", err)
		return
	}
	for fno, result := range taskResults {
		fmt.Printf("FNO: %d, RESULT: %v , STATUS: %v\n", fno, result.Result(), result.Error())
	}
	// Output:
	// err: fno: 3, TASK3 err
}

// JustNotBad 展示了非最佳的使用案例，包括，多任务创建、任务执行、结果收集，错误处理等
func ExampleTaskGroup_justNotBad() {
	tasks := []*taskgroup.Task{
		taskgroup.NewTask(1, task1ReturnFailWrapper(1, false), false),
		taskgroup.NewTask(2, task2ReturnSuccessWrapper(2, false), true),
		taskgroup.NewTask(3, task3ReturnFailWrapper(2, false), false),
		nil,
		nil,
	}

	taskResults, err := taskgroup.NewTaskGroup(nil).AddTask(tasks...).Run()
	if err != nil {
		fmt.Printf("err: %+v\n", err)
		return
	}
	for fno, result := range taskResults {
		fmt.Printf("FNO: %d, RESULT: %v , STATUS: %v\n", fno, result.Result(), result.Error())
	}
	// Unordered output:
	// FNO: 3, RESULT: TASK3: The data is 928 , STATUS: fno: 2, TASK3 err
	// FNO: 1, RESULT: 1127 , STATUS: fno: 1, TASK1 err
	// FNO: 2, RESULT: {1112 mlee} , STATUS: <nil>
}

// Abnormal 展示了异常的使用案例，包括，多任务创建、任务执行、结果收集，错误处理等
func ExampleTaskGroup_abnormal() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("err: %+v\n", r)
		}
	}()

	tasks := []*taskgroup.Task{
		taskgroup.NewTask(1, task1ReturnFailWrapper(1, false), true),
		taskgroup.NewTask(2, task2ReturnSuccessWrapper(2, false), true),
		taskgroup.NewTask(2, task3ReturnFailWrapper(3, false), true),
		nil,
		nil,
	}

	taskResults, err := taskgroup.NewTaskGroup(taskgroup.WithWorkerNums(uint32(len(tasks)))).AddTask(tasks...).Run()
	if err != nil {
		fmt.Printf("err: %+v\n", err)
		return
	}
	for fno, result := range taskResults {
		fmt.Printf("FNO: %d, RESULT: %v , STATUS: %v\n", fno, result.Result(), result.Error())
	}
	// Output:
	// err: AddTask: Already have the same Task 2
}
