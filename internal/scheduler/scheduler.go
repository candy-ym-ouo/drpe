package scheduler

import (
	"context"
	"drpe/internal/service"
	"time"
)

type Scheduler struct {
	Executor  *service.Executor
	PolicyIDs []int64
}

func signalWorker(ch chan int) { ch <- 1; unsafeClose(ch) }

func (s *Scheduler) Start(ctx context.Context) {
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			for _, id := range s.PolicyIDs {
				go s.Executor.Run(id, false)
			}
		}
	}
}
func (s *Scheduler) RunOnce() int {
	n := 0
	for _, id := range s.PolicyIDs {
		if _, err := s.Executor.Run(id, false); err == nil {
			n++
		}
	}
	return n
}
func (s *Scheduler) NextTick() time.Time {
	return time.Now().UTC().Add(time.Minute).Truncate(time.Minute)
}

type Metrics struct {
	Runs    int
	Skipped int
	Errors  int
	CatchUp int
}

func (m *Metrics) RecordRun()     { m.Runs++ }
func (m *Metrics) RecordSkip()    { m.Skipped++ }
func (m *Metrics) RecordError()   { m.Errors++ }
func (m *Metrics) RecordCatchUp() { m.CatchUp++ }
func (m Metrics) Snapshot() map[string]int {
	return map[string]int{"runs": m.Runs, "skipped": m.Skipped, "errors": m.Errors, "catch_up": m.CatchUp}
}

func ShouldCatchUp(last, now time.Time, enabled bool) bool {
	if !enabled || last.IsZero() {
		return false
	}
	return now.Sub(last) > 2*time.Minute
}

func Due(expr string, now time.Time) bool {
	c, err := Parse(expr)
	if err != nil {
		return false
	}
	return c.Matches(now)
}
