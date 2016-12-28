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

Semaphore

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
````


## Copyright, License & Contributors

### Submitting a Pull Request

1. Fork it.
2. Open a [Pull Request](https://github.com/o1egl/syncx/pulls)
3. Enjoy a refreshing Diet Coke and wait

SyncX is released under the MIT license. See [LICENSE](LICENSE)