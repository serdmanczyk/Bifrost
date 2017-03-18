package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/serdmanczyk/bifrost"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

const (
	defaultNumWorkers    int           = 10
	defaultNumJobs       int           = 10000
	defaultJobDuration   time.Duration = time.Microsecond
	defaultJobExpiration time.Duration = time.Minute * 5
)

var (
	numWorkers  = flag.Int("workers", defaultNumWorkers, "number of workers to spawn")
	numJobs     = flag.Int("jobs", defaultNumJobs, "number of jobs to create")
	jobDuration = flag.Int64("jobduration", int64(defaultJobDuration), "How long jobs last (time.Sleep")
	jobExpiry   = flag.Int64("expiration", int64(defaultJobExpiration), "How long until a finished job is purged")
	report      = flag.Bool("report", false, "Print a report of all complete jobs")
)

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	flag.Parse()
	elapsed := func() func() time.Duration {
		start := time.Now()
		return func() time.Duration {
			return time.Now().Sub(start)
		}
	}()

	var wg sync.WaitGroup
	jobIDs := make([]int, 0, *numJobs)

	dispatcher := bifrost.NewDispatcher(
		bifrost.Workers(*numWorkers),
		bifrost.JobExpiry(time.Duration(*jobExpiry)),
	)

	log.Printf("initialized %s\n", elapsed())

	createJob := func(jobNum uint) func(*bifrost.Job) {
		return func(j *bifrost.Job) {
			j.Log(fmt.Sprintf("%d: running", jobNum))
			time.Sleep(time.Duration(*jobDuration))
			j.Log(fmt.Sprintf("%d: stopped", jobNum))
			wg.Done()
			return
		}
	}

	// internally dispatcher job id's increase monotonically
	// from zero, so we can safely assume sudoJobID correlates
	for sudoJobID := 0; sudoJobID < *numJobs; sudoJobID++ {
		wg.Add(1)
		go func(id uint) {
			dispatcher.Queue(bifrost.JobRunnerFunc(createJob(id)))
		}(uint(sudoJobID))

		jobIDs = append(jobIDs, sudoJobID)
	}
	log.Printf("queued %s\n", elapsed())

	done := make(chan bool)
	go func() {
		wg.Wait()
		dispatcher.Stop()
		done <- true
		close(done)
	}()

	func() {
		for {
			select {
			case <-time.After(time.Millisecond * 200):
				if *report {
					// Grab a random job and print its status
					id := jobIDs[rand.Intn(len(jobIDs))]
					job, err := dispatcher.Status(uint(id))
					if err != nil {
						// ignore, job may have been purged
						continue
					}
					json.NewEncoder(os.Stdout).Encode(&job)
				}
			case <-done:
				return
			}
		}
	}()
	log.Printf("done %s\n", elapsed())
}