package bifrost

import (
	"time"
)

// DispatcherOpt is a function type for
// configuring new Dispatchers
type DispatcherOpt func(*Dispatcher)

// Workers sets the number of workers to spawn
func Workers(numWorkers int) DispatcherOpt {
	return func(d *Dispatcher) {
		d.numWorkers = numWorkers
	}
}

// JobExpiry sets the duration after a job completes
// to maintain the job's info for querying.  After
// expiry elapses, the job info will be purged.
//	Note: JobExpiry only controls the purge of FINISHED jobs.
//	There is currently not a provision for stopping a running
//	job that has continued to run beyond its expected duration.
func JobExpiry(expiry time.Duration) DispatcherOpt {
	return func(d *Dispatcher) {
		d.jobExpiry = expiry
	}
}
