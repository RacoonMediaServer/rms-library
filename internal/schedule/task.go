package schedule

import (
	"context"
	"time"
)

type runPolicy int

const (
	runInOrder runPolicy = iota
	runImmediately
	runAt
	runAfter
	runIdle
)

type OpResult int

const (
	OpResultDone OpResult = iota
	OpResultRetry
	OpResultRetryAfter
)

type Result struct {
	Result OpResult
	After  time.Duration
}

type ExecuteFn func(ctx context.Context) Result

type Task struct {
	Group string
	Fn    ExecuteFn

	run runPolicy
	dur time.Duration
	tm  time.Time

	timeout time.Duration

	scheduledAt time.Time
}

func (t *Task) Immediately() *Task {
	t.run = runImmediately
	return t
}

func (t *Task) After(d time.Duration) *Task {
	t.run = runAfter
	t.dur = d
	return t
}

func (t *Task) At(tm time.Time) *Task {
	t.run = runAt
	t.tm = tm
	return t
}

func (t *Task) WithTimeout(timeout time.Duration) *Task {
	t.timeout = timeout
	return t
}

func (t *Task) WhenIdle() *Task {
	t.run = runIdle
	return t
}
