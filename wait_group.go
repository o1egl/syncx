package syncx

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"
)

type Status int

const (
	// StatusIdle means that WG did not run yet
	StatusIdle Status = iota
	// StatusSuccess means successful execution of all tasks
	StatusSuccess
	// StatusTimeout means that job was broken by timeout
	StatusTimeout
	// StatusCanceled means that WG was canceled
	StatusCanceled
	// StatusError means that job was broken by error in one task (if stopOnError is true)
	StatusError
)

// WaitgroupFunc func
type WaitgroupFunc func() error

// AdvancedWaitGroup enhanced wait group struct
type AdvancedWaitGroup struct {
	sync.RWMutex
	context     context.Context
	stack       []WaitgroupFunc
	stopOnError bool
	status      Status
	errors      []error
}

// SetTimeout defines timeout for all tasks
func (wg *AdvancedWaitGroup) SetTimeout(t time.Duration) *AdvancedWaitGroup {
	wg.Lock()
	ctx := context.Background()
	if wg.context != nil {
		ctx = wg.context
	}
	wg.context, _ = context.WithTimeout(ctx, t)
	wg.Unlock()
	return wg
}

// SetContext defines Context
func (wg *AdvancedWaitGroup) SetContext(t context.Context) *AdvancedWaitGroup {
	wg.Lock()
	wg.context = t
	wg.Unlock()
	return wg
}

// SetStopOnError stops wiatgroup if any task returns error
func (wg *AdvancedWaitGroup) SetStopOnError(b bool) *AdvancedWaitGroup {
	wg.Lock()
	wg.stopOnError = b
	wg.Unlock()
	return wg
}

// Add adds new tasks into waitgroup
func (wg *AdvancedWaitGroup) Add(funcs ...WaitgroupFunc) *AdvancedWaitGroup {
	wg.Lock()
	wg.stack = append(wg.stack, funcs...)
	wg.Unlock()
	return wg
}

// Start runs tasks in separate goroutines and waits for their completion
func (wg *AdvancedWaitGroup) Start() *AdvancedWaitGroup {
	wg.Lock()
	defer wg.Unlock()
	wg.status = StatusSuccess

	if taskCount := len(wg.stack); taskCount > 0 {
		failed := make(chan error, taskCount)
		done := make(chan bool, taskCount)

	StarterLoop:
		for _, f := range wg.stack {
			// check if context is canceled
			select {
			case <-wg.doneChannel():
				break StarterLoop
			default:
			}

			go func(f WaitgroupFunc, failed chan<- error, done chan<- bool) {
				// Handle panic and pack it into stdlib error
				defer func() {
					if r := recover(); r != nil {
						buf := make([]byte, 1000)
						runtime.Stack(buf, false)
						failed <- errors.New(fmt.Sprintf("Panic handeled\n%v\n%s", r, string(buf)))
					}
				}()

				if err := f(); err != nil {
					failed <- err
				} else {
					done <- true
				}
			}(f, failed, done)
		}

	ForLoop:
		for taskCount > 0 {
			select {
			case err := <-failed:
				wg.errors = append(wg.errors, err)
				taskCount--
				if wg.stopOnError {
					wg.status = StatusError
					break ForLoop
				}
			case <-done:
				taskCount--
			case <-wg.doneChannel():
				if _, ok := wg.context.Deadline(); ok {
					wg.status = StatusTimeout
				} else {
					wg.status = StatusCanceled
				}
				break ForLoop
			}
		}
	}

	return wg
}

func (wg *AdvancedWaitGroup) doneChannel() <-chan struct{} {
	if wg.context != nil {
		return wg.context.Done()
	}
	return nil
}

// Reset performs cleanup task queue and reset state
func (wg *AdvancedWaitGroup) Reset() {
	wg.Lock()
	wg.stack = nil
	wg.stopOnError = false
	wg.status = StatusIdle
	wg.errors = nil
	wg.context = nil
	wg.Unlock()
}

// LastError returns last error that caught by execution process
func (wg *AdvancedWaitGroup) LastError() error {
	wg.RLock()
	defer wg.RUnlock()
	if l := len(wg.errors); l > 0 {
		return wg.errors[l-1]
	}
	return nil
}

// AllErrors returns all errors that caught by execution process
func (wg *AdvancedWaitGroup) AllErrors() []error {
	wg.RLock()
	defer wg.RUnlock()
	return wg.errors
}

// Status returns result state
func (wg *AdvancedWaitGroup) Status() Status {
	wg.RLock()
	defer wg.RUnlock()
	return wg.status
}
