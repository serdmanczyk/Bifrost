package bifrost

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// for godoc
func ExampleDispatcher() {
	stdoutWriter := json.NewEncoder(os.Stdout)
	dispatcher := NewDispatcher(
		Workers(4),
		JobExpiry(time.Millisecond),
	)

	// Queue a job func
	tracker := dispatcher.QueueFunc(func() error {
		time.Sleep(time.Microsecond)
		return nil
	})

	// Queue a 'JobRunner'
	dispatcher.Queue(JobRunnerFunc(func() error {
		time.Sleep(time.Microsecond)
		return nil
	}))

	// Print out incomplete status
	status := tracker.Status()
	stdoutWriter.Encode(&status)
	// {"ID":0,"Complete":false,"Start":"2017-03-23T21:51:27.140681968-07:00"}

	// wait on completion
	<-tracker.Done()
	// Status is now complete
	status = tracker.Status()
	stdoutWriter.Encode(&status)
	// {"ID":0,"Complete":true,"Success":true,"Start":"2017-03-23T21:51:27.140681968-07:00","Finish":"2017-03-23T21:51:27.140830827-07:00"}

	// Queue a job that will 'fail'
	tracker = dispatcher.QueueFunc(func() error {
		time.Sleep(time.Microsecond)
		return fmt.Errorf("Failed")
	})

	// Show failure status
	<-tracker.Done()
	status = tracker.Status()
	stdoutWriter.Encode(&status)
	// {"ID":2,"Complete":true,"Success":false,"Error":"Failed","Start":"2017-03-23T21:51:27.141026625-07:00","Finish":"2017-03-23T21:51:27.141079871-07:00"}

	// Query for a job's status.
	tracker, _ = dispatcher.Status(tracker.ID())
	status = tracker.Status()
	stdoutWriter.Encode(&status)
	// {"ID":2,"Complete":true,"Success":false,"Error":"Failed","Start":"2017-03-23T21:51:27.141026625-07:00","Finish":"2017-03-23T21:51:27.141079871-07:00"}

	// Show all jobs
	jobs := dispatcher.GetJobs()
	stdoutWriter.Encode(jobs)
	// [{"ID":2,"Complete":true,"Success":false,"Error":"Failed","Start":"2017-03-23T21:51:27.141026625-07:00","Finish":"2017-03-23T21:51:27.141079871-07:00"},{"ID":0,"Complete":true,"Success":true,"Start":"2017-03-23T21:51:27.140681968-07:00","Finish":"2017-03-23T21:51:27.140830827-07:00"},{"ID":1,"Complete":true,"Success":true,"Start":"2017-03-23T21:51:27.140684331-07:00","Finish":"2017-03-23T21:51:27.140873087-07:00"}]

	// wait for jobs to be purged
	time.Sleep(time.Millisecond * 5)

	// should now be empty
	jobs = dispatcher.GetJobs()
	stdoutWriter.Encode(jobs)
	// []

	dispatcher.Stop()
}

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
	jobfunc := func() (err error) {
		atomic.AddInt64(&a, iter)
		wg.Done()
		return
	}

	wg.Add(1)
	job := d.QueueFunc(jobfunc)
	status := job.Status()
	assert.Equal(t, status.Complete, false, "Job should not be complete")
	assert.Equal(t, status.Success, false, "Job should not be successful (hasn't run)")
	assert.Equal(t, status.Error, "", "Error should be blank (not run yet)")

	<-job.Done()
	<-job.Done()
	status = job.Status()
	assert.Equal(t, status.Complete, true, "Job should be complete")
	assert.Equal(t, status.Success, true, "Job be successful")
	assert.Equal(t, status.Error, "", "Error should be blank (succesful)")

	wg.Add(1)
	job = d.QueueFunc(jobfunc)
	status = job.Status()
	assert.Equal(t, status.Complete, false, "Job should not be complete")
	assert.Equal(t, status.Success, false, "Job should not be successful (hasn't run)")
	assert.Equal(t, status.Error, "", "Error should be blank (not run yet)")

	for i := 0; int64(i) < numJobs-2; i++ {
		wg.Add(1)
		d.Queue(JobRunnerFunc(jobfunc))
	}
	wg.Wait()
	d.Stop()

	expectedValue := (start + numJobs*iter)
	if a != expectedValue {
		t.Errorf("Expected final a value %d, got %d", expectedValue, a)
	}

	status = job.Status()
	assert.Equal(t, status.Complete, true, "Job should be complete")
	assert.Equal(t, status.Success, true, "Job should be successful")
	assert.Equal(t, status.Error, "", "Error should be blank (no errors)")
}
