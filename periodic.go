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
	"time"
)

// Periodic worker type
type PeriodicWorker struct {
	controller
}

// New return a new Periodic worker.
// Passing a function to execute and the time duration
// between each function call.
// The return worker is in Paused state.
func NewPeriodic(fn func(), timeout time.Duration) *PeriodicWorker {
	w := new(PeriodicWorker)
	w.ctl = make(chan State)

	waitGroup.Add(1)

	go func() {
		var state State = Paused

		defer waitGroup.Done()

		waitTimer := time.NewTimer(timeout)
		waitTimer.Stop()
		waitChan := waitTimer.C

	loop:
		for {
			if state == Running {
				waitTimer = time.NewTimer(timeout)
				waitChan = waitTimer.C
			}
			select {
			case state = <-w.ctl:
				switch state {
				case Paused:
					waitTimer.Stop() // Avoid the timer from firing
					goto loop
				case Shutdown:
					return
				}
			case <-waitChan:
			}
			fn() // Call the worker task
		}
	}()
	return w
}

// Start the worker
func (w *PeriodicWorker) Start() {
	w.ctl <- Running
}

// Pause the worker
func (w *PeriodicWorker) Pause() {
	w.ctl <- Paused
}

// Shutdown the worker. It can't be start again.
func (w *PeriodicWorker) Shutdown() {
	w.ctl <- Shutdown
}
