package schedule

import (
	"container/list"
	"time"
)

type queue struct {
	o *list.List
	t *list.List
	i *list.List
}

func newQueue() queue {
	return queue{
		o: list.New(),
		t: list.New(),
		i: list.New(),
	}
}

func (q queue) push(t *Task) {
	switch t.run {
	case runInOrder:
		q.o.PushBack(t)

	case runImmediately:
		q.o.PushFront(t)

	case runAfter:
		t.scheduledAt = time.Now().Add(t.dur)
		q.scheduleTask(t)

	case runAt:
		t.scheduledAt = t.tm
		q.scheduleTask(t)

	case runIdle:
		q.i.PushBack(t)
	}
}

func (q queue) pop(now time.Time) *Task {
	cur := q.t.Front()
	if cur != nil {
		t := cur.Value.(*Task)
		if t.scheduledAt.Before(now) {
			q.t.Remove(cur)
			return t
		}
	}

	cur = q.o.Front()
	if cur != nil {
		t := cur.Value.(*Task)
		q.o.Remove(cur)
		return t
	}

	cur = q.i.Front()
	if cur != nil {
		t := cur.Value.(*Task)
		q.i.Remove(cur)
		return t
	}

	return nil
}

func (q queue) scheduleTask(t *Task) {
	for cur := q.t.Front(); cur != nil; cur = cur.Next() {
		curTask := cur.Value.(*Task)
		if curTask.scheduledAt.After(t.scheduledAt) {
			q.t.InsertBefore(t, cur)
			return
		}
	}
	q.t.PushBack(t)
}

func (q queue) removeByGroup(group string) {
	fn := func(t *Task) bool {
		return t.Group == group
	}

	qRemove(q.o, fn)
	qRemove(q.t, fn)
	qRemove(q.i, fn)

}

func qRemove(q *list.List, fn func(*Task) bool) {
	for cur := q.Front(); cur != nil; {
		t := cur.Value.(*Task)
		if fn(t) {
			next := cur.Next()
			q.Remove(cur)
			cur = next
		} else {
			cur = cur.Next()
		}
	}
}
