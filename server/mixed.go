package server

import (
	"context"
	"fmt"
	"github.com/hylent/sf/logger"
	"github.com/soheilhy/cmux"
	"net"
	"sync"
)

type Mixed struct {
	ServerList []Server
}

func (x *Mixed) Match(cm cmux.CMux) net.Listener {
	return cm.Match(cmux.Any())
}

func (x *Mixed) Serve(ctx context.Context, listener net.Listener) error {
	switch len(x.ServerList) {
	case 0:
		return fmt.Errorf("mixed_server_empty")
	case 1:
		return x.ServerList[0].Serve(ctx, listener)
	}

	cm := cmux.New(listener)
	wg := new(sync.WaitGroup)
	errStrList := make([]string, len(x.ServerList)+1)

	for sIndex, s := range x.ServerList {
		sIndex := sIndex
		s := s
		l := s.Match(cm)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.Serve(context.TODO(), l); err != nil {
				errStrList[sIndex] = fmt.Sprintf("[%T]%s", err, err.Error())
			}
		}()
	}

	go func() {
		<-ctx.Done()
		cm.Close()
	}()

	if err := cm.Serve(); err != nil {
		errStrList[len(x.ServerList)] = fmt.Sprintf("[%T]%s", err, err.Error())
	}
	wg.Wait()

	log.Debug("mixed_server_failures", logger.M{
		"errStrList": errStrList,
	})

	return nil
}
