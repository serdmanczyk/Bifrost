package bifrost

import (
	"fmt"
	"sync"
	"time"
)

const (
	defaultNumWorkers int           = 4
	defaultJobExpiry  time.Duration = time.Minute * 5
	jobChanMargin     int           = 2
	jobMapCapMargin   int           = 5
)

// NewDispatcher create a new Dispatcher.
func NewDispatcher(opts ...DispatcherOpt) *Dispatcher {
	d := &Dispatcher{
		numWorkers:  defaultNumWorkers,
		jobExpiry:   defaultJobExpiry,
		workerQueue: make(chan chan *Job),
		stop:        make(chan bool),
	}

	for _, option := range opts {
		option(d)
	}

	d.workers = make([]*worker, 0, d.numWorkers)
	d.jobchan = make(chan *Job, d.numWorkers*jobChanMargin)
	d.jobsMap = make(map[uint]*Job, d.numWorkers*jobMapCapMargin)

	for i := 0; i < d.numWorkers; i++ {
		worker := newWorker(d.workerQueue)
		worker.start()
		d.workers = append(d.workers, worker)
	}
	d.start()
	return d
}

// Dispatcher is used to maintain and delegate work
// via Jobs to delegate workers.
type Dispatcher struct {
	workerQueue chan chan *Job
	jobchan     chan *Job
	stop        chan bool
	numWorkers  int
	currID      uint
	jobExpiry   time.Duration
	workers     []*worker
	jobsMap     map[uint]*Job
	mapLock     sync.RWMutex
}

func (d *Dispatcher) start() {
	auditTicker := time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case j := <-d.jobchan:
				go func(j *Job) {
					worker := <-d.workerQueue
					worker <- j
				}(j)
			case <-auditTicker.C:
				go d.jobMapAudit()
			case <-d.stop:
				d.stopWorkers()
				d.stop <- true
				close(d.stop)
				return
			}
		}
	}()
}

func (d *Dispatcher) stopWorkers() {
	stopChans := make([]chan bool, 0, len(d.workers))

	// signal all workers to finish what they're doing
	for _, worker := range d.workers {
		worker.stop <- true
		stopChans = append(stopChans, worker.stop)
	}
	// wait for workers to finish
	for _, stopChan := range stopChans {
		<-stopChan
	}
}

// Stop signals all workers to stop running their current
// jobs, waits for them to finish, then returns.
func (d *Dispatcher) Stop() {
	d.stop <- true
	<-d.stop
}

func (d *Dispatcher) jobMapAudit() {
	remIDs := make([]uint, 0, jobMapCapMargin)

	d.mapLock.RLock()
	for id, job := range d.jobsMap {
		finished, _, stopTime := job.Done()

		if finished && time.Now().Sub(stopTime) > d.jobExpiry {
			remIDs = append(remIDs, id)
		}
	}
	d.mapLock.RUnlock()

	if len(remIDs) > 0 {
		d.mapLock.Lock()
		for _, id := range remIDs {
			delete(d.jobsMap, id)
		}
		d.mapLock.Unlock()
	}
}

// Queue takes any implementer of the JobRunner interface
// and schedules it to be run via a worker.
func (d *Dispatcher) Queue(j JobRunner) *Job {
	d.mapLock.Lock()
	defer d.mapLock.Unlock()

	job := newJob(j, d.currID)
	d.jobsMap[job.ID] = job
	d.currID++
	d.jobchan <- job

	return job
}

// GetJobs returns all currently managed jobs
func (d *Dispatcher) GetJobs() (jobs []*Job) {
	d.mapLock.RLock()
	defer d.mapLock.RUnlock()

	for _, job := range d.jobsMap {
		jobs = append(jobs, job)
	}

	return jobs
}

// Status returns the Job for the given jobID.  The Job
// can be used to determine completion, success, and to
// view messages logged to the job.
func (d *Dispatcher) Status(jobID uint) (*Job, error) {
	d.mapLock.RLock()
	defer d.mapLock.RUnlock()

	job, ok := d.jobsMap[jobID]
	if !ok {
		return job, fmt.Errorf("Job %d not found", jobID)
	}

	return job, nil
}
