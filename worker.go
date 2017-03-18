package bifrost

import "time"

func newWorker(workQueue chan chan *Job) *worker {
	return &worker{
		workQueue: workQueue,
		work:      make(chan *Job),
		stop:      make(chan bool),
	}
}

type worker struct {
	workQueue chan chan *Job
	work      chan *Job
	stop      chan bool
}

func (w *worker) start() {
	go func() {

		for {
			select {
			case w.workQueue <- w.work:
				job := <-w.work
				job.run()
			case <-w.stop:
				w.stop <- true
				close(w.stop)
				return
			case <-time.After(time.Millisecond):
				continue
			}
		}
	}()
}
