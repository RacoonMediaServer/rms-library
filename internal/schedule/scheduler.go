package schedule

import (
	"context"
	"sync"
	"time"
)

const maxNotifications = 1000
const tickInterval = 10 * time.Second
const maxTaskTimeout = 10 * time.Minute

type Scheduler struct {
	ctx    context.Context
	cancel context.CancelFunc

	wg sync.WaitGroup

	notifies chan struct{}

	mu            sync.Mutex
	q             queue
	running       *Task
	cancelRunning bool
}

func New() *Scheduler {
	s := Scheduler{
		notifies: make(chan struct{}, maxNotifications),
		q:        newQueue(),
	}

	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.process()
	}()

	return &s
}

func (s *Scheduler) process() {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.notifies:
			s.processQueue()
		case <-ticker.C:
			s.processQueue()
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Scheduler) processQueue() {
	for {
		now := time.Now()
		s.mu.Lock()
		t := s.q.pop(now)
		s.running = t
		s.mu.Unlock()

		if t == nil {
			return
		}

		s.run(t)
	}
}

func (s *Scheduler) Cancel(group string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.q.removeByGroup(group)
	if s.running != nil && s.running.Group == group {
		s.cancelRunning = true
	}
}

func (s *Scheduler) Add(t *Task) bool {
	if t == nil || t.Fn == nil {
		return false
	}

	s.mu.Lock()
	s.q.push(t)
	s.mu.Unlock()

	s.notifies <- struct{}{}
	return true
}

func (s *Scheduler) run(t *Task) {
	var zero time.Duration
	timeout := maxTaskTimeout
	if t.timeout != zero {
		timeout = t.timeout
	}

	ctx, cancel := context.WithTimeout(s.ctx, timeout)
	defer cancel()

	result := t.Fn(ctx)

	switch result.Result {
	case OpResultRetry:
		if t.dur != 0 {
			t.dur = time.Second
		}
		t.dur *= 2
		t.scheduledAt = time.Now().Add(t.dur)

	case OpResultRetryAfter:
		t.scheduledAt = time.Now().Add(result.After)
		t.dur = result.After
	}

	s.mu.Lock()
	if result.Result != OpResultDone && !s.cancelRunning {
		s.q.scheduleTask(t)
	}
	s.running = nil
	s.cancelRunning = false
	s.mu.Unlock()
}

func (s *Scheduler) Stop() {
	s.cancel()
	s.wg.Wait()
	close(s.notifies)
}
