package taskgroup_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"time"
	_ "unsafe" // For go:linkname (编译)指令

	"github.com/mlee-msl/taskgroup"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func getRandomNumWithMod(mod int) int {
	return rand.Int() % mod
}

// ---------------------------------------------------
// ----------------------功能测试----------------------
// ---------------------------------------------------
// 结合白盒测试进行黑盒测试或灰盒测试
// 正常情况的功能测试（模糊测试）
//
// go test -gcflags="all=-N -l" -run=^$ -fuzz=FuzzTaskGroupRun_normal [-fuzztime=60s] ./...
func FuzzTaskGroupRun_normal(f *testing.F) {
	var (
		taskTotalMax        = 500 // 任务数数量上限
		taskKeyMax          = len(taskFuncSet)
		probabilityValueMax = 100
		successFlagMax      = 80 // `80%`的概率设置任务为必须执行成功
		taskStatusMax       = 95 // `95%`的概率任务不为空
	)

	testCases := []struct {
		workerNums           uint32
		taskTotalSeed        uint32 // 执行任务总数
		probabilityValueSeed uint32
	}{
		{0, 0, 2},
		{1, 1, 20},
		{3, 4, 30},
		{10, 40, 5},
		{30, 100, 102},
	}
	// 补充已知测试用例（白盒测试）
	// seed corpus for the fuzz test
	for _, testCase := range testCases {
		f.Add(testCase.workerNums, testCase.taskTotalSeed, testCase.probabilityValueSeed)
	}

	// 通过以下转化，将需要待执行的任务集合的[无限输入空间]转化为[有限输入空间]
	f.Fuzz(func(t *testing.T, workerNums, taskTotalSeed, probabilityValueSeed uint32) {
		var (
			taskTotal             = taskTotalSeed % uint32(taskTotalMax)
			tasks                 = make([]*taskgroup.Task, 0, taskTotal)
			taskFuncNos           = make([]taskFuncNo, 0, taskTotal)
			taskFuncSetStates     = make([]bool, 0, taskTotal)
			taskNoToTaskFuncNoMap = make(map[int]taskFuncNo, taskTotal)
		)
		for i := 0; i < int(taskTotal); i++ {
			var (
				taskFuncNo       = taskFuncNo(getRandomNumWithMod(taskKeyMax))
				probabilityValue = int(probabilityValueSeed) % probabilityValueMax
				taskFuncSetState = probabilityValue < successFlagMax
			)
			taskFuncNos = append(taskFuncNos, taskFuncNo)
			taskFuncSetStates = append(taskFuncSetStates, taskFuncSetState)
			taskNoToTaskFuncNoMap[i] = taskFuncNo
			if probabilityValue > taskStatusMax {
				tasks = append(tasks, nil)
				continue
			}
			tasks = append(tasks, taskgroup.NewTask(uint32(i), taskFuncSet[taskFuncNo], taskFuncSetState))
		}
		results, err := taskgroup.NewTaskGroup(taskgroup.WithWorkerNums(workerNums)).AddTask(tasks...).Run()
		if !checkTaskGroupRun(input{tasks, taskNoToTaskFuncNoMap, taskFuncNos, taskFuncSetStates}, output{results, err}) {
			t.Errorf("tasks=%+v, result=%+v, err=%+v", tasks, results, err)
		}
	})
}

type taskFuncNo byte

const (
	taskFuncNo0 taskFuncNo = iota
	taskFuncNo1
	taskFuncNo2
	taskFuncNo3
	taskFuncNo4
	taskFuncNo5
)

var (
	taskFuncSet = map[taskFuncNo]taskgroup.TaskFunc{
		taskFuncNo0: nil, // 空任务方法
		taskFuncNo1: task1ReturnFailWrapper(1, false),
		taskFuncNo2: task2ReturnSuccessWrapper(2, false),
		taskFuncNo3: task3ReturnFailWrapper(3, false),
		taskFuncNo4: task4ReturnSuccessWrapper(4, false),
		taskFuncNo5: task5ReturnFailWrapper(5, false),
	}
	taskFuncExecStates = map[taskFuncNo]bool{
		taskFuncNo0: true,  // 任务返回成功
		taskFuncNo1: false, // 任务返回失败
		taskFuncNo2: true,
		taskFuncNo3: false,
		taskFuncNo4: true,
		taskFuncNo5: false,
	}
)

type input struct {
	tasks                 []*taskgroup.Task
	taskNoToTaskFuncNoMap map[int]taskFuncNo
	taskFuncNos           []taskFuncNo
	taskFuncSetStates     []bool
}

type output struct {
	results map[uint32]*taskgroup.TaskResult
	err     error
}

func checkTaskGroupRun(i input, o output) bool {
	if l1, l2, l3, l4 := func() (taskLen, taskNoToTaskFuncNoMapLen, taskFuncNosLen, taskFuncSetStatesLen int) {
		return len(i.tasks), len(i.taskNoToTaskFuncNoMap), len(i.taskFuncNos), len(i.taskFuncSetStates)
	}(); l1 != l2 || l1 != l3 || l1 != l4 {
		return false
	}

	var execTaskNums int // 真正执行的任务数量
	for index, task := range i.tasks {
		if task == nil || taskFuncSet[i.taskFuncNos[index]] == nil {
			continue
		}

		execTaskNums++
		// 任务设置为必须成功，任务执行返回失败，则，任务组需要返回失败
		if i.taskFuncSetStates[index] && !taskFuncExecStates[i.taskFuncNos[index]] && o.err == nil {
			return false
		}
	}
	// 任务组执行返回失败，则，执行结果集不可信
	if o.err != nil {
		return true
	}

	// 任务组执行成功，则返回的任务的结果总数与执行的任务数量需保持一致
	if len(o.results) != execTaskNums {
		return false
	}
	for _, result := range o.results {
		taskFuncNo, has := i.taskNoToTaskFuncNoMap[int(result.FNO())]
		if !has {
			return false
		}
		if taskFuncExecStates[taskFuncNo] && result.Error() != nil || !taskFuncExecStates[taskFuncNo] && result.Error() == nil {
			return false
		}
	}
	return true
}

// 异常情况的功能测试
//
// go test -gcflags="all=-N -l" -run=^$ -fuzz=TestTaskGroupRun_abnormal [-fuzztime=60s] ./...
func TestTaskGroupRun_abnormal(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("No panic")
			return
		}
		t.Logf("err: %+v", r)
	}()

	tasks := []*taskgroup.Task{
		taskgroup.NewTask(1, task1ReturnFailWrapper(1, true), true),
		taskgroup.NewTask(2, task2ReturnSuccessWrapper(2, true), true),
		taskgroup.NewTask(2, task3ReturnFailWrapper(3, true), true),
	}

	_, _ = new(taskgroup.TaskGroup).AddTask(tasks...).Run()
}

//go:linkname RearrangeTasks github.com/mlee-msl/taskgroup.rearrangeTasks
func RearrangeTasks([]*taskgroup.Task)

func Test_rearrangeTasks(t *testing.T) {
	var (
		task1 = taskgroup.NewTask(1, nil, false)
		task2 = taskgroup.NewTask(2, func() (interface{}, error) { return nil, nil }, true)
		task3 = taskgroup.NewTask(2, func() (interface{}, error) { return nil, nil }, true)
	)

	testCases := []struct {
		tasks []*taskgroup.Task

		expected [][]*taskgroup.Task
	}{
		{
			[]*taskgroup.Task{nil, nil}, [][]*taskgroup.Task{{nil, nil}},
		},
		{
			[]*taskgroup.Task{task1, task2, task3},
			[][]*taskgroup.Task{{task2, task3, task1}, {task3, task2, task1}},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			RearrangeTasks(testCase.tasks)
			for _, excepted := range testCase.expected {
				if reflect.DeepEqual(testCase.tasks, excepted) {
					return
				}
			}
			t.Errorf("tasks=%+v, expected=%+v", testCase.tasks, testCase.expected)
		})
	}
}

//go:linkname AdjustWorkerNums github.com/mlee-msl/taskgroup.adjustWorkerNums
func AdjustWorkerNums(uint32, uint32) uint32

// go test -gcflags="all=-N -l" -run=^$ -fuzz=Fuzz_adjustWorkerNums -fuzztime=10s ./...
func Fuzz_adjustWorkerNums(f *testing.F) {
	addCases := []struct {
		workerNums, taskNums uint32
	}{
		{
			0, 1,
		},
		{
			0, 100,
		},
		{
			1, 10,
		},
		{
			2, 1000,
		},
	}
	// 补充已知测试用例（白盒测试）
	// seed corpus for the fuzz test
	for _, addCase := range addCases {
		f.Add(addCase.workerNums, addCase.taskNums)
	}

	f.Fuzz(func(t *testing.T, workerNums, taskNums uint32) {
		expected := workerNums
		if minPerWorkerTaskNums := (taskNums / 4) + 1; workerNums < minPerWorkerTaskNums {
			expected = minPerWorkerTaskNums
		} else if workerNums > taskNums {
			expected = taskNums
		}
		if got := AdjustWorkerNums(workerNums, taskNums); got != expected {
			t.Errorf("workerNums=%v, taskNums=%v, got=%v, expected=%v", workerNums, taskNums, got, expected)
		}
	})
}

// go test -run=^$ -fuzz=. -fuzztime=10s /...
func FuzzIf(f *testing.F) {
	f.Fuzz(func(t *testing.T, cond bool, a, b int) {
		expected := a
		if !cond {
			expected = b
		}
		if got := taskgroup.If(cond, a, b); got != expected {
			// t.Helper()
			t.Errorf("cond=%v, a=%v, b=%v, got=%v, expected=%v", cond, a, b, got, expected)
		}
	})
}

// ----------------------模拟的待执行的任务----------------------
type jobFunc func() error

// task1ReturnFail 模拟的并发任务1,且返回失败
func task1ReturnFail(fno uint32, isPrint bool) (interface{}, error) {
	const taskFlag = "is running TASK1(Return Fail)"
	if isPrint {
		fmt.Printf("fno: %d, %s\n", fno, taskFlag)
	}
	simulateIO(strings.Repeat(taskFlag, 10))
	return 1127, fmt.Errorf("fno: %d, TASK1 err", fno)
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
		a: 1112,
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
	return fmt.Sprintf("TASK3: The data is %d", 928), fmt.Errorf("fno: %d, TASK3 err", fno)
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

	time.Sleep(time.Duration(getRandomNumWithMod(10)) * time.Millisecond)
}
