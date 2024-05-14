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

// 功能测试
func TestTaskGroup(t *testing.T) {
	var (
		tg taskgroup.TaskGroup

		tasks = []*taskgroup.Task{
			taskgroup.NewTask(1, task1, true),
			taskgroup.NewTask(2, task2, true),
			taskgroup.NewTask(3, task3, true),
		}
	)

	taskResult, err := tg.SetWorkerNums(4).AddTask(tasks...).Run()
	fmt.Printf("**************TaskGroup************\n%+v, %+v\n", taskResult, err)
	for fno, result := range taskResult {
		fmt.Printf("FNO: %d, RESULT: %v , STATUS: %v\n", fno, result.Result(), result.Error())
	}
}

func TestTaskGroupBoundary(t *testing.T) {
	var (
		tg taskgroup.TaskGroup

		tasks = []*taskgroup.Task{}
	)

	taskResult, err := tg.SetWorkerNums(4).AddTask(tasks...).Run()
	fmt.Printf("**************TaskGroup************\n%+v, %+v\n", taskResult, err)
	for fno, result := range taskResult {
		fmt.Printf("FNO: %d, RESULT: %v , STATUS: %v\n", fno, result.Result(), result.Error())
	}
}

func TestTaskGroupAbnormal(t *testing.T) {
	var (
		tg taskgroup.TaskGroup

		tasks = []*taskgroup.Task{
			taskgroup.NewTask(1, task1, true),
			nil,
			taskgroup.NewTask(2, task2, false),
			taskgroup.NewTask(2, task1, true),
		}
	)

	taskResult, err := tg.SetWorkerNums(4).AddTask(tasks...).Run()
	fmt.Printf("**************TaskGroup************\n%+v, %+v\n", taskResult, err)
	for fno, result := range taskResult {
		fmt.Printf("FNO: %d, RESULT: %v , STATUS: %v\n", fno, result.Result(), result.Error())
	}
}

func getRandomNum(mod int) int {
	return rand.Int() % mod
}

func task1() (interface{}, error) {
	const taskFlag = "running TASK1"
	fmt.Println(taskFlag)
	simulateMemIO(taskFlag)
	return getRandomNum(2e3), fmt.Errorf("%s err", taskFlag)
}

type task2Struct struct {
	a int
	b string
}

func task2() (interface{}, error) {
	const taskFlag = "running TASK2"
	fmt.Println(taskFlag)
	simulateMemIO(taskFlag)
	return task2Struct{
		a: getRandomNum(1e1),
		b: "mlee",
	}, nil
}

func task3() (interface{}, error) {
	const taskFlag = "running TASK3"
	fmt.Println(taskFlag)
	simulateMemIO(taskFlag)
	return fmt.Sprintf("%s: The data is %d", taskFlag, getRandomNum(12)), fmt.Errorf("%s err", taskFlag)
}

type task4Struct struct {
	field0 task2Struct
	field1 uint32
	field2 string
	field3 []task2Struct
	field4 map[string]*task2Struct
	field5 *string
}

func task4() (interface{}, error) {
	const taskFlag = "TASK4"
	// fmt.Println(taskFlag)
	simulateMemIO(taskFlag)
	var field5 = "mleeeeeeeeeeeeeee"
	return task4Struct{
		field0: task2Struct{10, "ok"},
		field1: 1024,
		field2: "12",
		field3: []task2Struct{{12, "mlee1"}, {122, "mlee2"}, {1222, "mlee3"}},
		field4: map[string]*task2Struct{"a1": {10, "@@"}, "a2": {111, "##"}},
		field5: &field5,
	}, fmt.Errorf("%s err", taskFlag)
}

func simulateMemIO(s string) {
	var buf bytes.Buffer
	_, _ = buf.WriteString(s)
	_ = buf.String()
}

// -----------性能测试-----------
func BenchmarkTaskGroupZero(b *testing.B) {
	tasks := buildTestCaseData(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var tg taskgroup.TaskGroup
		_, _ = tg.SetWorkerNums(2).AddTask(tasks...).Run()
	}
}

func BenchmarkTaskGroupLow(b *testing.B) {
	tasks := buildTestCaseData(3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var tg taskgroup.TaskGroup
		_, _ = tg.SetWorkerNums(2).AddTask(tasks...).Run()
	}
}

func BenchmarkTaskGroupNormal(b *testing.B) {
	tasks := buildTestCaseData(8)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var tg taskgroup.TaskGroup
		_, _ = tg.SetWorkerNums(2).AddTask(tasks...).Run()
	}
}

func BenchmarkTaskGroupMedium(b *testing.B) {
	tasks := buildTestCaseData(15)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var tg taskgroup.TaskGroup
		_, _ = tg.SetWorkerNums(2).AddTask(tasks...).Run()
	}
}

func BenchmarkTaskGroupHigh(b *testing.B) {
	tasks := buildTestCaseData(40)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var tg taskgroup.TaskGroup
		_, _ = tg.SetWorkerNums(2).AddTask(tasks...).Run()
	}
}

func BenchmarkTaskGroupExtremelyHigh(b *testing.B) {
	tasks := buildTestCaseData(200)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var tg taskgroup.TaskGroup
		_, _ = tg.SetWorkerNums(2).AddTask(tasks...).Run()
	}
}

var taskSet = []func() (interface{}, error){task1, task2, task3, task4}

func buildTestCaseData(taskNums uint32) []*taskgroup.Task {
	if taskNums == 0 {
		return nil
	}
	tasks := make([]*taskgroup.Task, 0, taskNums)
	for i := 1; i <= int(taskNums); i++ {
		tasks = append(tasks, taskgroup.NewTask(uint32(getRandomNum(1e10)), taskSet[getRandomNum(len(taskSet))], true))
	}
	return tasks
}

// Typical 展示了典型的使用案例，包括，多任务创建、任务执行、结果收集，错误处理等
func ExampleTaskGroup_typical() {
	var (
		tg taskgroup.TaskGroup

		tasks = []*taskgroup.Task{
			taskgroup.NewTask(1, task1, true),
			taskgroup.NewTask(2, task2, false),
			taskgroup.NewTask(3, task3, true),
		}
	)

	taskResults, err := tg.SetWorkerNums(6).AddTask(tasks...).Run()
	if err != nil {
		fmt.Printf("err: %+v", err)
		return
	}
	for fno, result := range taskResults {
		fmt.Printf("FNO: %d, RESULT: %v , STATUS: %v\n", fno, result.Result(), result.Error())
	}
}