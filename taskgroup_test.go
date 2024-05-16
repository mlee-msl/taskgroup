package taskgroup_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/mlee-msl/taskgroup"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type jobFunc func() error

// ---------------------------------------------------
// ----------------------功能测试----------------------
// ---------------------------------------------------
func TestTaskGroupEarlyReturn(t *testing.T) {
	tg := taskgroup.NewTaskGroup(taskgroup.WithWorkerNums(20))
	tasks := []*taskgroup.Task{
		taskgroup.NewTask(0, task1ReturnFailWrapper(0, true), true),
		taskgroup.NewTask(1, task1ReturnFailWrapper(1, true), false),
		taskgroup.NewTask(2, task2ReturnSuccessWrapper(2, true), true),
		taskgroup.NewTask(3, task3ReturnFailWrapper(3, true), true),
		taskgroup.NewTask(4, task4ReturnSuccessWrapper(4, true), false),
		taskgroup.NewTask(5, task5ReturnFailWrapper(5, true), true),
		taskgroup.NewTask(6, task3ReturnFailWrapper(6, true), true),
		taskgroup.NewTask(7, task2ReturnSuccessWrapper(7, true), true),
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
			taskgroup.NewTask(1, task1ReturnFailWrapper(1, true), true),
			taskgroup.NewTask(2, task2ReturnSuccessWrapper(2, true), true),
			taskgroup.NewTask(3, task3ReturnFailWrapper(3, true), true),
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
			taskgroup.NewTask(1, task1ReturnFailWrapper(1, true), false),
			taskgroup.NewTask(2, task2ReturnSuccessWrapper(2, true), true),
			taskgroup.NewTask(3, task3ReturnFailWrapper(3, true), false),
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
			taskgroup.NewTask(1, task1ReturnFailWrapper(1, true), true),
			nil,
			taskgroup.NewTask(2, task2ReturnSuccessWrapper(2, true), false),
			taskgroup.NewTask(2, task1ReturnFailWrapper(2, true), true),
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
			taskgroup.NewTask(1, task1ReturnFailWrapper(1, true), true),
			nil,
			taskgroup.NewTask(2, task2ReturnSuccessWrapper(2, true), false),
			taskgroup.NewTask(2, task1ReturnFailWrapper(2, true), true),
		}
	)
	_, _ = tg.AddTask(tasks...).Run()
}

func getRandomNum(mod int) int {
	return rand.Int() % mod
}

// task1ReturnFail 模拟的并发任务1,且返回失败
func task1ReturnFail(fno uint32, isPrint bool) (interface{}, error) {
	const taskFlag = "is running TASK1(Return Fail)"
	if isPrint {
		fmt.Printf("fno: %d, %s\n", fno, taskFlag)
	}
	simulateIO(strings.Repeat(taskFlag, 10))
	return getRandomNum(2e3), fmt.Errorf("fno: %d, TASK1 err", fno)
}

func task1ReturnFailWrapper(fno uint32, isPrint bool) taskgroup.TaskFunc {
	return func() (interface{}, error) {
		return task1ReturnFail(fno, isPrint)
	}
}

func task1ReturnFailWrapperForErrGroup(fno uint32, isPrint bool) jobFunc {
	return func() error {
		_, err := task1ReturnFail(fno, isPrint)
		return err
	}
}

type task2Struct struct {
	a int
	b string
}

// task2ReturnSuccess 模拟的并发任务2，且返回成功
func task2ReturnSuccess(fno uint32, isPrint bool) (interface{}, error) {
	const taskFlag = "is running TASK2(Return Success)"
	if isPrint {
		fmt.Printf("fno: %d, %s\n", fno, taskFlag)
	}
	simulateIO(strings.Repeat(taskFlag, 10))
	return task2Struct{
		a: getRandomNum(1e1),
		b: "mlee",
	}, nil
}

func task2ReturnSuccessWrapper(fno uint32, isPrint bool) taskgroup.TaskFunc {
	return func() (interface{}, error) {
		return task2ReturnSuccess(fno, isPrint)
	}
}

func task2ReturnSuccessWrapperForErrGroup(fno uint32, isPrint bool) jobFunc {
	return func() error {
		_, err := task2ReturnSuccess(fno, isPrint)
		return err
	}
}

// task3ReturnFail 模拟的并发任务3，且返回失败
func task3ReturnFail(fno uint32, isPrint bool) (interface{}, error) {
	const taskFlag = "is running TASK3(Return Fail)"
	if isPrint {
		fmt.Printf("fno: %d, %s\n", fno, taskFlag)
	}
	simulateIO(strings.Repeat(taskFlag, 10))
	return fmt.Sprintf("TASK3: The data is %d", getRandomNum(12)), fmt.Errorf("fno: %d, TASK3 err", fno)
}

func task3ReturnFailWrapper(fno uint32, isPrint bool) taskgroup.TaskFunc {
	return func() (interface{}, error) {
		return task3ReturnFail(fno, isPrint)
	}
}

func task3ReturnFailWrapperForErrGroup(fno uint32, isPrint bool) jobFunc {
	return func() error {
		_, err := task3ReturnFail(fno, isPrint)
		return err
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
func task4ReturnSuccess(fno uint32, isPrint bool) (interface{}, error) {
	const taskFlag = "is running TASK4(Return Success)"
	if isPrint {
		fmt.Printf("fno: %d, %s\n", fno, taskFlag)
	}
	simulateIO(strings.Repeat(taskFlag, 10))
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

func task4ReturnSuccessWrapper(fno uint32, isPrint bool) taskgroup.TaskFunc {
	return func() (interface{}, error) {
		return task4ReturnSuccess(fno, isPrint)
	}
}

func task4ReturnSuccessWrapperForErrGroup(fno uint32, isPrint bool) jobFunc {
	return func() error {
		_, err := task4ReturnSuccess(fno, isPrint)
		return err
	}
}

// task5ReturnFail 模拟的并发任务5,且返回失败
func task5ReturnFail(fno uint32, isPrint bool) (interface{}, error) {
	const taskFlag = "is running TASK5(Return Fail)"
	if isPrint {
		fmt.Printf("fno: %d, %s\n", fno, taskFlag)
	}
	simulateIO(strings.Repeat(taskFlag, 10))
	var field5 = "mleeeeeeeeeeeeeee"
	return task4Struct{
		field0: task2Struct{1024, "futu"},
		field1: 1024,
		field2: "00700",
		field3: []task2Struct{{12, "mlee1"}, {122, "mlee2"}, {1222, "mlee3"}},
		field4: map[string]*task2Struct{"a1": {11, "@@"}, "a2": {111, "##"}, "a3": {1111, "$$"}},
		field5: &field5,
	}, fmt.Errorf("fno: %d, TASK5 err", fno)
}

func task5ReturnFailWrapper(fno uint32, isPrint bool) taskgroup.TaskFunc {
	return func() (interface{}, error) {
		return task5ReturnFail(fno, isPrint)
	}
}

func task5ReturnFailWrapperForErrGroup(fno uint32, isPrint bool) jobFunc {
	return func() error {
		_, err := task5ReturnFail(fno, isPrint)
		return err
	}
}

func simulateIO(s string) {
	var buf bytes.Buffer
	_, _ = buf.WriteString(s)
	_ = buf.String()

	time.Sleep(time.Duration(getRandomNum(800)) * time.Millisecond)
}

func TestIf(t *testing.T) {
	// min 函数
	a, b := 1, 2
	fmt.Println(taskgroup.If(a <= b, a, b).(int) == a)

	// max 函数
	fmt.Println(taskgroup.If(a >= b, a, b).(int) == b)
}
