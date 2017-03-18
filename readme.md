 ᛉ Bifröst
------
[![Go Report Card](https://goreportcard.com/badge/github.com/serdmanczyk/Bifrost)](https://goreportcard.com/report/github.com/serdmanczyk/Bifrost)
[![GoDoc](https://godoc.org/github.com/serdmanczyk/Bifrost?status.svg)](https://godoc.org/github.com/serdmanczyk/Bifrost)
[![blog](https://img.shields.io/badge/readMy-blog-green.svg)](http://serdmanczyk.github.io)

![gofrost](repo/gofrost.png "Freyr")

## What?

If you've read the blog posts [Handling 1 Million Requests per Minute with Go](http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/) or [Writing worker queues, in Go](http://nesv.github.io/golang/2014/02/25/worker-queues-in-go.html) this will look very familiar.  The main machinery in Bifrost is basically identical to the functionality described in those blog posts, but with a couple added features I wanted for my project.

Added Features:

    - Generic queuable jobs: a job is any struct or function that accepts a *Job as an argument.
                             The passed in Job struct can also be used to log errors, etc.
    - Graceful shutdown: when dispatcher is stopped, waits for running jobs to complete.
    - Tracking: queued jobs are given an ID and held in an internal map
                (leverageable via an API to give a user job status).
    - Cleanup: completed jobs are cleaned up after a configurable amount of time.

Lacks (might add these later):

    - Lost jobs: if the dispatcher is stopped before all jobs are sent to a worker, unsent jobs may be ignored.
    - Errant jobs: jobs taking longer than expected cannot be cancelled.
    - Single process: this package does not include functionality to schedule jobs across multiple processes
                      via AMQP, gRPC, or otherwise.

For an example, see the [test](dispatcher_test.go) or example [command line app](examples/main.go).

Obligatory "not for use in production" but I do welcome feedback.

## Etymology

[Bifröst](https://en.wikipedia.org/wiki/Bifr%C3%B6st) (pronounce B-eye-frost popularly, or traditionally more like Beef-roast but not exactly like that you filthy American(/s)) is the bridge between the realms of Earth and Asgard (the heavens) in norse mythology.  The Futhark [ᛉ Elhaz/Algiz](https://en.wikipedia.org/wiki/Algiz) is seen as the symbol for Bifröst, or at least according to [this thing I Googled](http://vrilology.org/FUTHARK.htm).  The symbology intended is that jobs are dispatched between Earth (the queueing goroutine) to the heavens (the worker queue) i.e. I just needed a cool Norse thing to name my project after.  Not to be taken too seriously, but I do accept corrections/lessons from people who actually know what they're talking about.
