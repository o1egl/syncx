[![License](http://img.shields.io/:license-mit-blue.svg)](LICENSE)
[![GoDoc](https://godoc.org/github.com/o1egl/syncx?status.svg)](https://godoc.org/github.com/o1egl/syncx)
[![Build Status](http://img.shields.io/travis/o1egl/syncx.svg?style=flat-square)](https://travis-ci.org/o1egl/syncx)
[![Coverage Status](http://img.shields.io/coveralls/o1egl/syncx.svg?style=flat-square)](https://coveralls.io/r/o1egl/syncx)
# SyncX

SyncX is a GO library that extends standard sync package.

## Install

```
$ go get -u github.com/o1egl/syncx
```

## Usage
* [AdwancedWaitGroup](#adwancedwaitgroup)
* [Semaphore](#semaphore)

### AdwancedWaitGroup

AWG -  advanced version of wait group

Added features:

* thread safe
* integrated context.Context
* execution timeout
* ability to return errors
* panic handling from goroutines

```go
    wg := awg.AdvancedWaitGroup{}

    // Add first task
    wg.Add(func() error {
    	//Logic
    	return nil
    })

    // Add second task
    wg.Add(func() error {
    	//Another Logic
    	return nil
    })

    wg.Start()

    var err error
    // Taking one error make sense if you use *.SetStopOnError(true)* option - see below
    err = wg.SetStopOnError(true).Start().GetLastError()

    // Taking all errors
    var errs []error
    errs = wg.Start().GetAllErrors()    

```

Integrated with context.Context. It gives you ability to set timeouts and register cancel functions

```go
    // SetTimeout defines timeout for all tasks
    SetTimeout(t time.Duration)

    // SetContext defines Context
    SetContext(t context.Context)

    // SetStopOnError stops waitgroup if any task returns error
    SetStopOnError(b bool)

    // Add adds new tasks into waitgroup
    Add(funcs ...WaitgroupFunc)

    // // Start runs tasks in separate goroutines and waits for their completion
    Start()

    // Reset performs cleanup task queue and reset state
    Reset()

    // // LastError returns last error that caught by execution process
    LastError()

    // AllErrors returns all errors that caught by execution process
    AllErrors()

    // Status returns result state
    Status()
```


### Semaphore

```go
    // NewSemaphore returns new Semaphore instance
    NewSemaphore(10)

    //Acquire acquires one permit, if its not available the goroutine will block till its available or Context.Done() occurs.
    //You can pass context.WithTimeout() to support timeoutable acquire.
    Acquire(ctx context.Context)

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
```



## Copyright, License & Contributors

### Submitting a Pull Request

1. Fork it.
2. Open a [Pull Request](https://github.com/o1egl/syncx/pulls)
3. Enjoy a refreshing Diet Coke and wait

SyncX is released under the MIT license. See [LICENSE](LICENSE)
