package syncx

import (
	"context"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
	"time"
)

func TestNewSemaphore(t *testing.T) {
	_, err := NewSemaphore(0)
	assert.NotNil(t, err)

	s, err := NewSemaphore(10)
	assert.Nil(t, err)
	assert.NotNil(t, s)
	assert.Equal(t, 10, s.AvailablePermits())
}

func TestSemaphoreAcquire(t *testing.T) {
	s, _ := NewSemaphore(1)
	ctx := context.Background()

	s.Acquire(ctx)
	counter := 0
	start := make(chan struct{})
	end := make(chan struct{})
	go func() {
		<-start
		s.Acquire(ctx)
		counter++
		close(end)
	}()

	assert.Equal(t, 0, counter)
	start <- struct{}{}
	s.Release()
	<-end
	assert.Equal(t, 1, counter)
}

func TestSemaphoreAcquireMany(t *testing.T) {
	s, _ := NewSemaphore(4)
	ctx := context.Background()

	acquired, err := s.AcquireMany(ctx, 2)
	assert.Nil(t, err)
	assert.Equal(t, 2, acquired)
	assert.Equal(t, 2, s.AvailablePermits())

	err = s.ReleaseMany(2)
	assert.Nil(t, err)
	assert.Equal(t, 4, s.AvailablePermits())

	counter := 0
	steps := make(chan struct{})
	go func() {
		<-steps
		acquired, err := s.DrainPermits(ctx)
		assert.Nil(t, err)
		assert.Equal(t, 4, acquired)

		steps <- struct{}{}
		<-steps

		err = s.Acquire(ctx)
		counter++
		assert.Nil(t, err)

		s.Release()
		close(steps)
	}()

	steps <- struct{}{}

	<-steps
	assert.Equal(t, 0, s.AvailablePermits())
	assert.Equal(t, 0, counter)
	err = s.ReleaseMany(4)
	assert.Nil(t, err)
	steps <- struct{}{}

	<-steps
	assert.Equal(t, 4, s.AvailablePermits())
	assert.Equal(t, 1, counter)

	_, err = s.AcquireMany(ctx, -1)
	assert.NotNil(t, err)

	_, err = s.AcquireMany(ctx, 10)
	assert.NotNil(t, err)

	err = s.ReleaseMany(10)
	assert.NotNil(t, err)

	err = s.ReleaseMany(-1)
	assert.NotNil(t, err)

	s.DrainPermits(ctx)
	acquired, _ = s.DrainPermits(ctx)
	assert.Equal(t, 0, acquired)
}

func TestSemaphoreAcquireCancel(t *testing.T) {
	s, _ := NewSemaphore(4)
	d := 2 * time.Second
	ctx, _ := context.WithTimeout(context.Background(), d)

	s.DrainPermits(ctx)
	assert.Equal(t, 0, s.AvailablePermits())

	start := time.Now()
	err := s.Acquire(ctx)

	assert.NotNil(t, err)
	assert.True(t, math.Floor(time.Now().Sub(start).Seconds()) == d.Seconds())

	s.ReleaseMany(4)
	assert.Equal(t, 4, s.AvailablePermits())

	ctx, _ = context.WithTimeout(context.Background(), d)
	acquired, err := s.AcquireMany(ctx, 4)
	assert.Nil(t, err)
	assert.Equal(t, 4, acquired)
	assert.Equal(t, 0, s.AvailablePermits())

	start = time.Now()
	ctx, _ = context.WithTimeout(context.Background(), d)
	acquired, err = s.AcquireMany(ctx, 4)
	assert.NotNil(t, err)
	assert.Equal(t, 0, acquired)
	assert.True(t, math.Floor(time.Now().Sub(start).Seconds()) == d.Seconds())
}
