package bifrost

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDispatcher(t *testing.T) {
	const start int64 = 10
	const iter int64 = 10
	const numWorkers int = 10
	const numJobs int64 = 20
	var wg sync.WaitGroup

	d := NewDispatcher(
		Workers(numWorkers),
		JobExpiry(time.Second),
	)
	a := start
	jobfunc := JobRunnerFunc(func(j *Job) {
		newVal := atomic.AddInt64(&a, iter)
		j.Log(fmt.Sprintf("%d", newVal))
		wg.Done()
	})

	wg.Add(1)
	job := d.Queue(jobfunc)
	for i := 0; int64(i) < numJobs-1; i++ {
		wg.Add(1)
		d.Queue(jobfunc)
	}
	wg.Wait()
	d.Stop()

	expectedValue := (start + numJobs*iter)
	if a != expectedValue {
		t.Errorf("Expected final a value %d, got %d", expectedValue, a)
	}

	if len(job.Logs()) != 1 {
		t.Errorf("Expected 1 log statement got %d", len(job.Logs()))
	}
}
