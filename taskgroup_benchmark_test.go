package taskgroup_test

import (
	"context"
	"math/rand"
	"runtime"
	"testing"

	"github.com/mlee-msl/taskgroup"
	"golang.org/x/sync/errgroup"
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
	const taskNums = uint32(40)
	tasks := buildTestCaseData(taskNums)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		workerNums := uint32(runtime.NumCPU()) // 初步验证，协程数和逻辑`cpu`量一致时，性能表现最佳
		// workerNums := taskNums
		tg := taskgroup.NewTaskGroup(taskgroup.WithWorkerNums(workerNums))
		_, _ = tg.AddTask(tasks...).Run()
	}
}

func BenchmarkErrGroupHigh(b *testing.B) {
	const taskNums = 40

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		errGroup, _ := errgroup.WithContext(context.Background())
		for i := 1; i <= taskNums; i++ {
			errGroup.Go(taskSetForErrGroup[rand.Intn(len(taskSetForErrGroup))])
		}
		_ = errGroup.Wait()
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

var (
	taskSetForTaskGroup = []taskgroup.TaskFunc{
		task1ReturnFailWrapper(1, false),
		task2ReturnSuccessWrapper(2, false),
		task3ReturnFailWrapper(3, false),
		task4ReturnSuccessWrapper(4, false),
		task5ReturnFailWrapper(5, false),
	}

	taskSetForErrGroup = []jobFunc{
		task1ReturnFailWrapperForErrGroup(1, false),
		task2ReturnSuccessWrapperForErrGroup(2, false),
		task3ReturnFailWrapperForErrGroup(3, false),
		task4ReturnSuccessWrapperForErrGroup(4, false),
		task5ReturnFailWrapperForErrGroup(5, false),
	}
)

func buildTestCaseData(taskNums uint32) []*taskgroup.Task {
	if taskNums == 0 {
		return nil
	}
	tasks := make([]*taskgroup.Task, 0, taskNums)
	for i := 1; i <= int(taskNums); i++ {
		var mustSuccess bool
		if rand.Float64() < 0.8 { // `80%`的接口要求必须是成功的
			mustSuccess = true
		}
		tasks = append(tasks, taskgroup.NewTask(uint32(i), taskSetForTaskGroup[rand.Intn(len(taskSetForTaskGroup))], mustSuccess))
	}
	return tasks
}

// go test -benchmem -run=^$ -bench ^BenchmarkTaskGroup$ -cpu='1,2,4,8,16' [-benchtime=10x|-benchtime=1s] -count=5 .
// go test -benchmem -run=^$ -bench ^BenchmarkTaskGroup$ -cpu='1,2,4,8,16' -benchtime=10x -count=5 .
// go test -benchmem -run=^$ -bench ^BenchmarkTaskGroup$ -cpu='1,2,4,8,16' -benchtime=10x -count=5 -cpuprofile='cpu.pprof' .
// go test -benchmem -run=^$ -bench ^BenchmarkTaskGroup$ -cpu='1,2,4,8,16' -benchtime=10x -count=5 -memprofile='mem.pprof' .
// go test -benchmem -run=^$ -bench ^BenchmarkTaskGroup$ -cpu='1,2,4,8,16' -benchtime=10x -count=5 -blockprofile='block.pprof' .
// go test -benchmem -run=^$ -bench ^BenchmarkTaskGroup$ -cpu='1,2,4,8,16' -benchtime=10x -count=5 -mutexprofile='mutex.pprof' .
func BenchmarkTaskGroup(b *testing.B) {
	const taskNums = uint32(60)
	tasks := buildTestCaseData(taskNums)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		workerNums := uint32(runtime.NumCPU()) // 初步验证，协程数和逻辑`cpu`量一致时，性能表现最佳
		// workerNums := taskNums
		tg := taskgroup.NewTaskGroup(taskgroup.WithWorkerNums(workerNums))
		_, _ = tg.AddTask(tasks...).Run()
	}
}

// go test -benchmem -run=^$ -bench ^BenchmarkErrGroup$ [-benchtime=10x|-benchtime=1s -count=5 .
// go test -benchmem -run=^$ -bench ^BenchmarkErrGroup$ -cpu='1,2,4,8,16' -benchtime=10x -count=5 .
func BenchmarkErrGroup(b *testing.B) {
	const taskNums = 60

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		errGroup, _ := errgroup.WithContext(context.Background())
		for i := 1; i <= taskNums; i++ {
			errGroup.Go(taskSetForErrGroup[rand.Intn(len(taskSetForErrGroup))])
		}
		_ = errGroup.Wait()
	}
}
