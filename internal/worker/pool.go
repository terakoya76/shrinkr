// Package worker runs a bounded fan-out over Jobs. It never aborts on
// single-job failures — the caller inspects the returned Results.
package worker

import (
	"context"
	"sync"

	"github.com/terakoya76/shrinkr/internal/plan"
)

type Action int

const (
	ActionCompressed Action = iota
	ActionCopied
	ActionSkipped
	ActionFailed
)

func (a Action) String() string {
	switch a {
	case ActionCompressed:
		return "compressed"
	case ActionCopied:
		return "copied"
	case ActionSkipped:
		return "skipped"
	case ActionFailed:
		return "failed"
	default:
		return "unknown"
	}
}

type Result struct {
	Job     plan.Job
	Action  Action
	DstSize int64
	Reason  string
	Err     error
}

type Handler func(ctx context.Context, job plan.Job) Result

// Run executes handler for every job with at most workers concurrent goroutines.
// It streams every Result to onResult in the order jobs complete.
func Run(ctx context.Context, jobs []plan.Job, workers int, handler Handler, onResult func(Result)) {
	if workers < 1 {
		workers = 1
	}
	jobCh := make(chan plan.Job)
	resCh := make(chan Result)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobCh {
				select {
				case <-ctx.Done():
					return
				default:
				}
				resCh <- handler(ctx, job)
			}
		}()
	}

	go func() {
		defer close(jobCh)
		for _, j := range jobs {
			select {
			case <-ctx.Done():
				return
			case jobCh <- j:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(resCh)
	}()

	for r := range resCh {
		onResult(r)
	}
}
