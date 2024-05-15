package taskgroup_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/mlee-msl/taskgroup"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// ---------------------------------------------------
// ----------------------功能测试----------------------
// ---------------------------------------------------
func TestTaskGroupEarlyReturn(t *testing.T) {
	tg := taskgroup.NewTaskGroup(taskgroup.WithWorkerNums(20))
	tasks := []*taskgroup.Task{
		taskgroup.NewTask(0, task1ReturnFailWrapper(0), true),
		taskgroup.NewTask(1, task1ReturnFailWrapper(1), false),
		taskgroup.NewTask(2, task2ReturnSuccessWrapper(2), true),
		taskgroup.NewTask(3, task3ReturnFailWrapper(3), true),
		taskgroup.NewTask(4, task3ReturnFailWrapper(4), false),
		taskgroup.NewTask(5, task1ReturnFailWrapper(5), true),
		taskgroup.NewTask(6, task3ReturnFailWrapper(6), true),
		taskgroup.NewTask(7, task2ReturnSuccessWrapper(7), true),
	}
	taskResults, err := tg.AddTask(tasks...).Run()
	if err != nil {
		fmt.Printf("taskResults:%+v, err: %+v\n", taskResults, err)
		return
	}
	for fno, result := range taskResults {
		fmt.Printf("FNO: %d, RESULT: %v , STATUS: %v\n", fno, result.Result(), result.Error())
	}
}

// 功能测试
func TestTaskGroup(t *testing.T) {
	var (
		tg taskgroup.TaskGroup

		tasks = []*taskgroup.Task{
			taskgroup.NewTask(1, task1ReturnFailWrapper(1), true),
			taskgroup.NewTask(2, task2ReturnSuccessWrapper(2), true),
			taskgroup.NewTask(3, task3ReturnFailWrapper(3), true),
		}
	)

	taskResults, err := tg.AddTask(tasks...).Run()
	fmt.Printf("**************TaskGroup************\n%+v, %+v\n", taskResults, err)
	for fno, result := range taskResults {
		fmt.Printf("FNO: %d, RESULT: %v , STATUS: %v\n", fno, result.Result(), result.Error())
	}
}

func TestTaskGroupCtx(t *testing.T) {
	var (
		tg taskgroup.TaskGroup

		tasks = []*taskgroup.Task{
			taskgroup.NewTask(1, task1ReturnFailWrapper(1), false),
			taskgroup.NewTask(2, task2ReturnSuccessWrapper(2), true),
			taskgroup.NewTask(3, task3ReturnFailWrapper(3), false),
		}
	)

	taskResults, err := tg.AddTask(tasks...).Run()
	fmt.Printf("**************TaskGroup************\n%+v, %+v\n", taskResults, err)
	for fno, result := range taskResults {
		fmt.Printf("FNO: %d, RESULT: %v , STATUS: %v\n", fno, result.Result(), result.Error())
	}
}

func TestTaskGroupBoundary(t *testing.T) {
	var (
		tg *taskgroup.TaskGroup

		tasks = []*taskgroup.Task{}
	)
	tg = taskgroup.NewTaskGroup(taskgroup.WithWorkerNums(4))
	taskResults, err := tg.AddTask(tasks...).Run()
	fmt.Printf("**************TaskGroup************\n%+v, %+v\n", taskResults, err)
	for fno, result := range taskResults {
		fmt.Printf("FNO: %d, RESULT: %v , STATUS: %v\n", fno, result.Result(), result.Error())
	}
}

func TestTaskGroupAbnormal(t *testing.T) {
	var (
		tg *taskgroup.TaskGroup

		tasks = []*taskgroup.Task{
			taskgroup.NewTask(1, task1ReturnFailWrapper(1), true),
			nil,
			taskgroup.NewTask(2, task2ReturnSuccessWrapper(2), false),
			taskgroup.NewTask(2, task1ReturnFailWrapper(2), true),
		}
	)
	tg = taskgroup.NewTaskGroup()
	taskResults, err := tg.AddTask(tasks...).Run()
	fmt.Printf("**************TaskGroup************\n%+v, %+v\n", taskResults, err)
	for fno, result := range taskResults {
		fmt.Printf("FNO: %d, RESULT: %v , STATUS: %v\n", fno, result.Result(), result.Error())
	}
}

func TestTaskGroupError(t *testing.T) {
	var (
		tg *taskgroup.TaskGroup

		tasks = []*taskgroup.Task{
			taskgroup.NewTask(1, task1ReturnFailWrapper(1), true),
			nil,
			taskgroup.NewTask(2, task2ReturnSuccessWrapper(2), false),
			taskgroup.NewTask(2, task1ReturnFailWrapper(2), true),
		}
	)
	_, _ = tg.AddTask(tasks...).Run()
}

func getRandomNum(mod int) int {
	return rand.Int() % mod
}

// task1ReturnFail 模拟的并发任务1,且返回失败
func task1ReturnFail(fno uint32) (interface{}, error) {
	const taskFlag = "is running TASK1"
	fmt.Printf("fno: %d, %s\n", fno, taskFlag)
	simulateMemIO(taskFlag)
	return getRandomNum(2e3), fmt.Errorf("fno: %d, TASK1 err", fno)
}

func task1ReturnFailWrapper(fno uint32) taskgroup.TaskFunc {
	return func() (interface{}, error) {
		return task1ReturnFail(fno)
	}
}

type task2Struct struct {
	a int
	b string
}

// task2ReturnSuccess 模拟的并发任务2，且返回成功
func task2ReturnSuccess(fno uint32) (interface{}, error) {
	const taskFlag = "is running TASK2"
	fmt.Printf("fno: %d, %s\n", fno, taskFlag)
	simulateMemIO(taskFlag)
	return task2Struct{
		a: getRandomNum(1e1),
		b: "mlee",
	}, nil
}

func task2ReturnSuccessWrapper(fno uint32) taskgroup.TaskFunc {
	return func() (interface{}, error) {
		return task2ReturnSuccess(fno)
	}
}

// task3ReturnFail 模拟的并发任务3，且返回失败
func task3ReturnFail(fno uint32) (interface{}, error) {
	const taskFlag = "is running TASK3"
	fmt.Printf("fno: %d, %s\n", fno, taskFlag)
	simulateMemIO(taskFlag)
	return fmt.Sprintf("TASK3: The data is %d", getRandomNum(12)), fmt.Errorf("fno: %d, TASK3 err", fno)
}

func task3ReturnFailWrapper(fno uint32) taskgroup.TaskFunc {
	return func() (interface{}, error) {
		return task3ReturnFail(fno)
	}
}

type task4Struct struct {
	field0 task2Struct
	field1 uint32
	field2 string
	field3 []task2Struct
	field4 map[string]*task2Struct
	field5 *string
}

// task4ReturnSuccess 模拟的并发任务4,且返回成功
func task4ReturnSuccess(fno uint32) (interface{}, error) {
	const taskFlag = "is running TASK4"
	fmt.Printf("fno: %d, %s\n", fno, taskFlag)
	simulateMemIO(taskFlag)
	var field5 = "mleeeeeeeeeeeeeee"
	return task4Struct{
		field0: task2Struct{10, "ok"},
		field1: 1024,
		field2: "12",
		field3: []task2Struct{{12, "mlee1"}, {122, "mlee2"}, {1222, "mlee3"}},
		field4: map[string]*task2Struct{"a1": {10, "@@"}, "a2": {111, "##"}, "a3": {111, "$$"}},
		field5: &field5,
	}, nil
}

func task4ReturnSuccessWrapper(fno uint32) taskgroup.TaskFunc {
	return func() (interface{}, error) {
		return task4ReturnSuccess(fno)
	}
}

func simulateMemIO(s string) {
	var buf bytes.Buffer
	_, _ = buf.WriteString(s)
	_ = buf.String()
}
