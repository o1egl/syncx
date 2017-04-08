package syncx

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
	"time"
)

var testError = errors.New("Test error")

func slowFunc() error {
	time.Sleep(2 * time.Second)
	context.WithTimeout(context.Background(), 10*time.Second)
	return nil
}

func fastFunc() error {
	// do nothing
	return nil
}

func errorFunc() error {
	return testError
}

func panicFunc() error {
	panic("Test panic")
	return nil
}

// Test_AdvancedWaitGroup_Success test for success case
func Test_AdvancedWaitGroup_Success(t *testing.T) {
	var wg AdvancedWaitGroup

	wg.Add(fastFunc)
	wg.Add(fastFunc)
	wg.Add(slowFunc)
	wg.Add(slowFunc)

	start := time.Now()
	wg.Start()
	diff := time.Now().Sub(start).Nanoseconds()

	assert.True(t, diff >= (2*time.Second).Nanoseconds(), "AWG should wait all goroutines")
	assert.Equal(t, StatusSuccess, wg.Status())
	assert.NoError(t, wg.LastError())
	assert.Len(t, wg.AllErrors(), 0)
}

// Test_AdvancedWaitGroup_SuccessWithErrors test for success case
func Test_AdvancedWaitGroup_SuccessWithErrors(t *testing.T) {
	var wg AdvancedWaitGroup

	wg.Add(fastFunc)
	wg.Add(fastFunc)
	wg.Add(errorFunc)
	wg.Add(errorFunc)
	wg.Add(slowFunc)
	wg.Add(slowFunc)

	start := time.Now()
	wg.Start()
	diff := time.Now().Sub(start).Nanoseconds()

	assert.True(t, diff >= (2*time.Second).Nanoseconds(), "AWG should wait all goroutines")
	assert.Equal(t, StatusSuccess, wg.Status())
	assert.Error(t, testError, wg.LastError())
	assert.Len(t, wg.AllErrors(), 2)
}

// Test_AdvancedWaitGroup_Timeout test for timeout
func Test_AdvancedWaitGroup_Timeout(t *testing.T) {
	var wg AdvancedWaitGroup

	wg.Add(fastFunc)
	wg.Add(fastFunc)
	wg.Add(fastFunc)
	wg.Add(slowFunc)
	wg.Add(slowFunc)
	wg.Add(slowFunc)

	start := time.Now()
	wg.SetContext(context.Background()).SetTimeout(time.Nanosecond * 10).
		Start()
	diff := time.Now().Sub(start).Nanoseconds()

	assert.True(t, diff < (time.Second).Nanoseconds(), "AWG should be canceled immediately")
	assert.Equal(t, StatusTimeout, wg.Status(), "AWG should stops by timeout!")
}

// Test_AdvancedWaitGroup_Cancel test for cancel
func Test_AdvancedWaitGroup_Cancel(t *testing.T) {
	var wg AdvancedWaitGroup
	ctx, cancelFunc := context.WithCancel(context.Background())

	wg.Add(fastFunc)
	wg.Add(fastFunc)
	wg.Add(fastFunc)
	wg.Add(slowFunc)
	wg.Add(slowFunc)
	wg.Add(slowFunc)

	go func() {
		time.Sleep(200 * time.Millisecond)
		cancelFunc()
	}()

	start := time.Now()
	wg.SetContext(ctx).
		Start()
	diff := time.Now().Sub(start).Nanoseconds()

	assert.True(t, diff < (time.Second).Nanoseconds(), "AWG should be canceled immediately")
	assert.Equal(t, StatusCanceled, wg.Status(), "AWG status should be StatusCanceled")
}

// Test_AdvancedWaitGroup_DontExecCanceled test for don't run goroutines if cantext is canceled
func Test_AdvancedWaitGroup_DontExecCanceled(t *testing.T) {
	var wg AdvancedWaitGroup
	ctx, cancelFunc := context.WithCancel(context.Background())

	i := 0
	wg.Add(func() error {
		i++
		return nil
	})

	cancelFunc()

	wg.SetContext(ctx).
		Start()

	assert.Equal(t, StatusCanceled, wg.Status(), "AWG status should be StatusCanceled")
	assert.Equal(t, 0, i)
}

// Test_AdvancedWaitGroup_StopOnError test for error
func Test_AdvancedWaitGroup_StopOnError(t *testing.T) {
	var wg AdvancedWaitGroup

	wg.Add(errorFunc)
	wg.Add(slowFunc)

	start := time.Now()
	wg.SetStopOnError(true).
		Start()
	diff := time.Now().Sub(start).Nanoseconds()

	assert.True(t, diff < (time.Second).Nanoseconds(), "AWG should be canceled immediately")
	assert.Equal(t, testError, wg.LastError())
	assert.Equal(t, StatusError, wg.Status(), "AWG status should be StatusError!")
}

// Test_AdvancedWaitGroup_Panic test for success case
func Test_AdvancedWaitGroup_Panic(t *testing.T) {
	var wg AdvancedWaitGroup

	wg.Add(slowFunc)
	wg.Add(panicFunc)

	wg.SetStopOnError(true).
		Start()

	assert.Equal(t, StatusError, wg.Status())
	assert.Contains(t, wg.LastError().Error(), "Test panic")
}

// Test_AdvancedWaitGroup_Reset test for reset
func Test_AdvancedWaitGroup_Reset(t *testing.T) {
	var wg AdvancedWaitGroup

	wg.Add(fastFunc).
		SetTimeout(5 * time.Second).
		SetContext(context.Background()).
		SetStopOnError(true).
		Start()

	wg.Reset()

	assert.Zero(t, wg)
}

// Test_AdvancedWaitGroup_NoLeak tests for goroutines leaks
func Test_AdvancedWaitGroup_NoLeak(t *testing.T) {
	var wg AdvancedWaitGroup

	wg.Add(errorFunc)

	wg.SetStopOnError(true).
		Start()

	time.Sleep(3 * time.Second)

	numGoroutines := runtime.NumGoroutine()

	var wg2 AdvancedWaitGroup

	wg2.Add(errorFunc)
	wg2.Add(slowFunc)

	wg2.SetStopOnError(true).
		Start()

	time.Sleep(3 * time.Second)

	numGoroutines2 := runtime.NumGoroutine()

	if numGoroutines != numGoroutines2 {
		t.Fatalf("We leaked %d goroutine(s)", numGoroutines2-numGoroutines)
	}
}
