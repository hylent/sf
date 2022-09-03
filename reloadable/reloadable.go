package reloadable

import (
	"fmt"
	"github.com/hylent/sf/logger"
	"sync/atomic"
	"time"
	"unsafe"
)

var log = logger.NewLogger(nil, "github.com/hylent/sf/reloadable")

type Reloadable[T any] interface {
	Get() *T
}

type defaultReloadable[T any] struct {
	ptr unsafe.Pointer
	ch  chan *T
}

func (x *defaultReloadable[T]) Get() (data *T) {
	return (*T)(atomic.LoadPointer(&x.ptr))
}

func (x *defaultReloadable[T]) set(data *T) {
	atomic.StorePointer(&x.ptr, unsafe.Pointer(data))
}

func New[T any](initTimeout time.Duration, f func(ch chan<- *T)) (Reloadable[T], error) {
	x := &defaultReloadable[T]{
		ch: make(chan *T, 1),
	}

	go f(x.ch)

	select {
	case data := <-x.ch:
		if data == nil {
			return nil, fmt.Errorf("reloadable_init_nil")
		}
		x.set(data)
	case <-time.After(initTimeout):
		return nil, fmt.Errorf("reloadable_init_timeout")
	}

	go func() {
		for data := range x.ch {
			if data != nil {
				x.set(data)
			}
		}
	}()

	return x, nil
}

type Poller[T any] interface {
	InitTimeout() time.Duration
	Interval() time.Duration
	GetData() (int64, *T, error)
	IsOutdated(int64) bool
}

func FromPoller[T any](poller Poller[T]) (Reloadable[T], error) {
	f := func(ch chan<- *T) {
		var currentVersion int64

		// init
		{
			version, data, dataErr := poller.GetData()
			if dataErr != nil {
				log.Warn("reloadable_init_fail", logger.M{
					"err": dataErr.Error(),
				})
				ch <- nil
				return
			}
			currentVersion = version
			ch <- data
		}

		for range time.NewTicker(poller.Interval()).C {
			oldVersion := currentVersion
			startTp := time.Now()
			if poller.IsOutdated(oldVersion) {
				version, data, dataErr := poller.GetData()
				if dataErr != nil {
					log.Warn("reloadable_reload_fail", logger.M{
						"err":     dataErr.Error(),
						"version": oldVersion,
						"cost":    time.Now().Sub(startTp).Milliseconds(),
					})
					time.Sleep(time.Second)
					continue
				}
				currentVersion = version
				ch <- data
			}
			log.Debug("reloadable_reload_succeed", logger.M{
				"cost":     time.Now().Sub(startTp).Milliseconds(),
				"version":  oldVersion,
				"version2": currentVersion,
			})
		}
	}

	return New[T](poller.InitTimeout(), f)
}
