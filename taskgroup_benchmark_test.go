package taskgroup_test

import (
	"testing"

	"github.com/mlee-msl/taskgroup"
)

// ---------------------------------------------------
// ----------------------性能测试----------------------
// ---------------------------------------------------
func BenchmarkTaskGroupZero(b *testing.B) {
	tasks := buildTestCaseData(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var tg taskgroup.TaskGroup
		_, _ = tg.AddTask(tasks...).Run()
	}
}

func BenchmarkTaskGroupLow(b *testing.B) {
	tasks := buildTestCaseData(3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var tg taskgroup.TaskGroup
		_, _ = tg.AddTask(tasks...).Run()
	}
}

func BenchmarkTaskGroupNormal(b *testing.B) {
	tasks := buildTestCaseData(8)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var tg taskgroup.TaskGroup
		_, _ = tg.AddTask(tasks...).Run()
	}
}

func BenchmarkTaskGroupMedium(b *testing.B) {
	tasks := buildTestCaseData(15)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var tg taskgroup.TaskGroup
		_, _ = tg.AddTask(tasks...).Run()
	}
}

func BenchmarkTaskGroupHigh(b *testing.B) {
	tasks := buildTestCaseData(40)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var tg taskgroup.TaskGroup
		_, _ = tg.AddTask(tasks...).Run()
	}
}

func BenchmarkTaskGroupExtremelyHigh(b *testing.B) {
	tasks := buildTestCaseData(200)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var tg taskgroup.TaskGroup
		_, _ = tg.AddTask(tasks...).Run()
	}
}

var taskSet = []func() (interface{}, error){
	task1ReturnFailWrapper(1),
	task2ReturnSuccessWrapper(2),
	task3ReturnFailWrapper(3),
	task4ReturnSuccessWrapper(4)}

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
