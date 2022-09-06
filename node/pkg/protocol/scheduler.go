// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"container/list"
	statemachine "github.com/atomix/multi-raft-storage/node/pkg/statemachine2"
	"time"
)

func newScheduler() *stateMachineScheduler {
	return &stateMachineScheduler{
		scheduledTasks: list.New(),
		indexedTasks:   make(map[statemachine.Index]*list.List),
		time:           time.Now(),
	}
}

type stateMachineScheduler struct {
	scheduledTasks *list.List
	indexedTasks   map[statemachine.Index]*list.List
	time           time.Time
}

func (s *stateMachineScheduler) Time() time.Time {
	return s.time
}

func (s *stateMachineScheduler) Await(index statemachine.Index, f func()) statemachine.Timer {
	task := &indexTask{
		scheduler: s,
		index:     index,
		callback:  f,
	}
	task.schedule()
	return task
}

func (s *stateMachineScheduler) Delay(d time.Duration, f func()) statemachine.Timer {
	task := &timeTask{
		scheduler: s,
		time:      s.time.Add(d),
		callback:  f,
	}
	task.schedule()
	return task
}

func (s *stateMachineScheduler) Schedule(t time.Time, f func()) statemachine.Timer {
	task := &timeTask{
		scheduler: s,
		time:      t,
		callback:  f,
	}
	task.schedule()
	return task
}

// tick runs the scheduled time-based tasks
func (s *stateMachineScheduler) tick(time time.Time) {
	s.runScheduledTasks(time)
	s.time = time
}

// tock runs the scheduled index-based tasks
func (s *stateMachineScheduler) tock(index statemachine.Index) {
	s.runIndexedTasks(index)
}

func (s *stateMachineScheduler) runIndexedTasks(index statemachine.Index) {
	if tasks, ok := s.indexedTasks[index]; ok {
		elem := tasks.Front()
		for elem != nil {
			task := elem.Value.(*indexTask)
			task.run()
			elem = elem.Next()
		}
		delete(s.indexedTasks, index)
	}
}

// runScheduleTasks runs the scheduled tasks in the scheduler queue
func (s *stateMachineScheduler) runScheduledTasks(time time.Time) {
	element := s.scheduledTasks.Front()
	if element != nil {
		for element != nil {
			task := element.Value.(*timeTask)
			if task.isRunnable(time) {
				next := element.Next()
				s.scheduledTasks.Remove(element)
				s.time = task.time
				task.run()
				element = next
			} else {
				break
			}
		}
	}
}

// time-based task
type timeTask struct {
	scheduler *stateMachineScheduler
	callback  func()
	time      time.Time
	elem      *list.Element
}

func (t *timeTask) schedule() {
	if t.scheduler.scheduledTasks.Len() == 0 {
		t.elem = t.scheduler.scheduledTasks.PushBack(t)
	} else {
		element := t.scheduler.scheduledTasks.Back()
		for element != nil {
			time := element.Value.(*timeTask).time
			if element.Value.(*timeTask).time.UnixNano() < time.UnixNano() {
				t.elem = t.scheduler.scheduledTasks.InsertAfter(t, element)
				return
			}
			element = element.Prev()
		}
		t.elem = t.scheduler.scheduledTasks.PushFront(t)
	}
}

func (t *timeTask) isRunnable(time time.Time) bool {
	return time.UnixNano() > t.time.UnixNano()
}

func (t *timeTask) run() {
	t.callback()
}

func (t *timeTask) Cancel() {
	if t.elem != nil {
		t.scheduler.scheduledTasks.Remove(t.elem)
	}
}

// index-based task
type indexTask struct {
	scheduler *stateMachineScheduler
	callback  func()
	index     statemachine.Index
	elem      *list.Element
}

func (t *indexTask) schedule() {
	tasks, ok := t.scheduler.indexedTasks[t.index]
	if !ok {
		tasks = list.New()
		t.scheduler.indexedTasks[t.index] = tasks
	}
	t.elem = tasks.PushBack(t)
}

func (t *indexTask) run() {
	t.callback()
}

func (t *indexTask) Cancel() {
	if t.elem != nil {
		if tasks, ok := t.scheduler.indexedTasks[t.index]; ok {
			tasks.Remove(t.elem)
		}
	}
}
