// Copyright © 2016 Yves Blusseau
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package worker

import (
	"sync"
	"time"
)

// Possible worker states.
const (
	Running  = 0
	Paused   = 1
	Shutdown = 2
)

// State of a worker
type State int

// Worker type
type Worker struct {
	ctl chan State // The channel to control the worker
}

type task func()

var waitGroup sync.WaitGroup

// New return a new worker.
// Passing a function to execute and the time duration
// between each function call
func New(fn task, timeout time.Duration) *Worker {
	w := new(Worker)
	w.ctl = make(chan State)

	waitGroup.Add(1)

	ticker := time.NewTicker(timeout)
	tickerChan := ticker.C

	go func() {
		var state State
		defer waitGroup.Done()
		ticker.Stop() // Start with a paused worker
	loop:
		for {
			select {
			case state = <-w.ctl:
				switch state {
				case Running:
					ticker = time.NewTicker(timeout)
					tickerChan = ticker.C
				case Paused:
					ticker.Stop()
					goto loop
				case Shutdown:
					return
				}
			case <-tickerChan:
			}
			fn() // Call the worker task
		}
	}()
	return w
}

// WaitAllWorker is used to wait that all workers are stop
func WaitAllWorker() {
	waitGroup.Wait()
}

// Start the worker
func (w *Worker) Start() {
	w.ctl <- Running
}

// Pause the worker
func (w *Worker) Pause() {
	w.ctl <- Paused
}

// Shutdown the worker. It can't be start again.
func (w *Worker) Shutdown() {
	w.ctl <- Shutdown
}
