package util

import (
	"context"
	"github.com/google/uuid"
	"github.com/hylent/sf/logger"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func Terminated(ctx context.Context, f func(ctx context.Context)) <-chan struct{} {
	fCtx, cancelFunc := context.WithCancel(ctx)
	go func() {
		defer cancelFunc()
		terminated := make(chan os.Signal, 1)
		signal.Notify(terminated, syscall.SIGTERM, syscall.SIGINT)
		<-terminated
	}()
	return Finished(fCtx, f)
}

func Finished(ctx context.Context, f func(ctx context.Context)) <-chan struct{} {
	finished := make(chan struct{})
	go func() {
		f(ctx)
		close(finished)
	}()
	return finished
}

func RunTaskList(maxConcurrency int, taskList []func()) {
	if maxConcurrency < 2 {
		// run orderly
		for index := range taskList {
			taskList[index]()
		}
		return
	}

	if maxConcurrency >= len(taskList) {
		// run concurrently
		wg := sync.WaitGroup{}
		for index := range taskList {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				taskList[index]()
			}(index)
		}
		wg.Wait()
		return
	}

	// prepare group id
	groupId := uuid.NewString()

	// prepare queue
	indexQueue := make(chan int, len(taskList))
	wg := sync.WaitGroup{}

	// prepare consumers
	for concurrencyId := 0; concurrencyId < maxConcurrency; concurrencyId++ {
		wg.Add(1)
		go func(concurrencyId int) {
			defer wg.Done()
			for index := range indexQueue {
				taskList[index]()
				logger.Debug("task_done", logger.M{
					"groupId":       groupId,
					"concurrencyId": concurrencyId,
					"index":         index,
				})
			}
		}(concurrencyId)
	}

	// produce
	for index := 0; index < len(taskList); index++ {
		indexQueue <- index
	}

	// close queue to notify consumers
	close(indexQueue)

	// wait for everything done
	wg.Wait()
}
