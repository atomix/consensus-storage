// SPDX-FileCopyrightText: 2022-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package raft

import (
	"container/list"
	"github.com/atomix/multi-raft-storage/node/pkg/primitive"
	"time"
)

// Timer is a cancellable timer
type Timer = primitive.Timer

func newScheduler() *Scheduler {
	return &Scheduler{
		tasks:          list.New(),
		scheduledTasks: list.New(),
		time:           time.Now(),
	}
}

type Scheduler struct {
	tasks          *list.List
	scheduledTasks *list.List
	time           time.Time
}

func (s *Scheduler) Time() time.Time {
	return s.time
}

func (s *Scheduler) Run(f func()) {
	s.tasks.PushBack(f)
}

func (s *Scheduler) RunAfter(d time.Duration, f func()) Timer {
	task := &task{
		scheduler: s,
		time:      time.Now().Add(d),
		interval:  0,
		callback:  f,
	}
	s.schedule(task)
	return task
}

func (s *Scheduler) RepeatAfter(d time.Duration, i time.Duration, f func()) Timer {
	task := &task{
		scheduler: s,
		time:      time.Now().Add(d),
		interval:  i,
		callback:  f,
	}
	s.schedule(task)
	return task
}

func (s *Scheduler) RunAt(t time.Time, f func()) Timer {
	task := &task{
		scheduler: s,
		time:      t,
		interval:  0,
		callback:  f,
	}
	s.schedule(task)
	return task
}

func (s *Scheduler) RepeatAt(t time.Time, i time.Duration, f func()) Timer {
	task := &task{
		scheduler: s,
		time:      t,
		interval:  i,
		callback:  f,
	}
	s.schedule(task)
	return task
}

// runImmediateTasks runs the immediate tasks in the scheduler queue
func (s *Scheduler) runImmediateTasks() {
	task := s.tasks.Front()
	for task != nil {
		task.Value.(func())()
		task = task.Next()
	}
	s.tasks = list.New()
}

// runScheduleTasks runs the scheduled tasks in the scheduler queue
func (s *Scheduler) runScheduledTasks(time time.Time) {
	s.time = time
	element := s.scheduledTasks.Front()
	if element != nil {
		complete := list.New()
		for element != nil {
			task := element.Value.(*task)
			if task.isRunnable(time) {
				next := element.Next()
				s.scheduledTasks.Remove(element)
				s.time = task.time
				task.run()
				complete.PushBack(task)
				element = next
			} else {
				break
			}
		}

		element = complete.Front()
		for element != nil {
			task := element.Value.(*task)
			if task.interval > 0 {
				task.time = s.time.Add(task.interval)
				s.schedule(task)
			}
			element = element.Next()
		}
	}
}

// schedule schedules a task
func (s *Scheduler) schedule(t *task) {
	if s.scheduledTasks.Len() == 0 {
		t.element = s.scheduledTasks.PushBack(t)
	} else {
		element := s.scheduledTasks.Back()
		for element != nil {
			time := element.Value.(*task).time
			if element.Value.(*task).time.UnixNano() < time.UnixNano() {
				t.element = s.scheduledTasks.InsertAfter(t, element)
				return
			}
			element = element.Prev()
		}
		t.element = s.scheduledTasks.PushFront(t)
	}
}

// Scheduler task
type task struct {
	Timer
	scheduler *Scheduler
	interval  time.Duration
	callback  func()
	time      time.Time
	element   *list.Element
}

func (t *task) isRunnable(time time.Time) bool {
	return time.UnixNano() > t.time.UnixNano()
}

func (t *task) run() {
	t.callback()
}

func (t *task) Cancel() {
	if t.element != nil {
		t.scheduler.scheduledTasks.Remove(t.element)
	}
}
