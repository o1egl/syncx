package syncx

import (
	"context"
	"errors"
)

// Semaphore defines semaphore interface
type Semaphore interface {
	//Acquire acquires one permit, if its not available the goroutine will block till its available or Context.Done() occurs.
	//You can pass context.WithTimeout() to support timeoutable acquire.
	Acquire(ctx context.Context) error

	//AcquireMany is similar to Acquire() but for many permits
	//Returns successfully acquired permits.
	AcquireMany(ctx context.Context, n int) (int, error)

	//Release releases one permit
	Release()

	//ReleaseMany releases many permits
	ReleaseMany(n int) error

	//AvailablePermits returns number of available unacquired permits
	AvailablePermits() int

	//DrainPermits acquires all available permits and return the number of permits acquired
	DrainPermits() (int, error)
}

// NewSemaphore returns new Semaphore instance
func NewSemaphore(permits int) (*semaphore, error) {
	if permits < 1 {
		return nil, errors.New("invalid number of permits. Less than 1")
	}
	return &semaphore{
		channel: make(chan struct{}, permits),
	}, nil
}

type semaphore struct {
	channel chan struct{}
}

func (s *semaphore) Acquire(ctx context.Context) error {
	select {
	case s.channel <- struct{}{}:
		return nil
	case <-ctx.Done():
		return errors.New("acquire canceled")
	}

}

func (s *semaphore) AcquireMany(ctx context.Context, n int) (int, error) {
	if n < 0 {
		return 0, errors.New("acquir count coundn't be negative")
	}
	if n > s.totalPermits() {
		return 0, errors.New("too many requested permits")
	}
	acquired := 0
	for ; n > 0; n-- {
		select {
		case s.channel <- struct{}{}:
			acquired++
			continue
		case <-ctx.Done():
			return acquired, errors.New("acquire canceled")
		}

	}
	return acquired, nil
}

func (s *semaphore) AvailablePermits() int {
	return s.totalPermits() - len(s.channel)
}

func (s *semaphore) DrainPermits(ctx context.Context) (int, error) {
	n := s.AvailablePermits()
	if n > 0 {
		return s.AcquireMany(ctx, n)
	}
	return n, nil
}

func (s *semaphore) Release() {
	<-s.channel
}

func (s *semaphore) ReleaseMany(n int) error {
	if n < 0 {
		return errors.New("release count coundn't be negative")
	}
	if n > s.totalPermits() {
		return errors.New("too many requested releases")
	}
	for ; n > 0; n-- {
		s.Release()
	}
	return nil
}

func (s *semaphore) totalPermits() int {
	return cap(s.channel)
}
