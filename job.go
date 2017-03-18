package bifrost

import (
	"encoding/json"
	"sync"
	"time"
)

// JobRunner interface is the accepted type for a
// job to be dispatched in the background.
type JobRunner interface {
	Run(*Job)
}

// JobRunnerFunc is a type used to wrap pure functions
// accepting a *Job, allowing them to implement the
// JobRunner interface.
type JobRunnerFunc func(*Job)

// Run implements the JobRunner interface for a
// JobFunc.
func (f JobRunnerFunc) Run(job *Job) {
	f(job)
}

func newJob(j JobRunner, id uint) *Job {
	job := Job{
		ID:       id,
		logs:     make([]string, 0),
		start:    time.Now(),
		complete: false,
	}

	var once sync.Once
	job.run = func() {
		once.Do(func() {
			j.Run(&job)
			job.lock.Lock()
			defer job.lock.Unlock()
			job.complete = true
			job.finish = time.Now()
		})
	}

	return &job
}

// Job is used to wrap the execution of work
// distributed to workers in the Dispatcher
// queue.
type Job struct {
	ID       uint
	logs     []string
	lock     sync.RWMutex
	complete bool
	start    time.Time
	finish   time.Time
	run      func()
}

// Logs is used to retreive a job's logs.
func (j *Job) Logs() []string {
	j.lock.RLock()
	defer j.lock.RUnlock()

	return j.logs
}

// Log is used for recording any messages concerning
// the Job's execution e.g. failure or warning messages.
func (j *Job) Log(log string) {
	j.lock.Lock()
	defer j.lock.Unlock()

	j.logs = append(j.logs, log)
}

// Done returns whether the task has finished, and if
// so, also the finish time.  If the job is not finished,
// te finish time should be ignored.
func (j *Job) Done() (bool, time.Time, time.Time) {
	j.lock.RLock()
	defer j.lock.RUnlock()

	return j.complete, j.start, j.finish
}

// MarshalJSON is used to implement the encoding/json
// Marshaler interface.  Implemented manually to allow
// read/write locking.
func (j *Job) MarshalJSON() ([]byte, error) {
	j.lock.RLock()
	defer j.lock.RUnlock()

	return json.Marshal(struct {
		ID       uint      `json: "id"`
		Complete bool      `json: "complete"`
		Logs     []string  `json: "logs"`
		Start    time.Time `json: "start"`
		Finish   time.Time `json: "finish"`
	}{
		j.ID,
		j.complete,
		j.logs,
		j.start,
		j.finish,
	})
}
