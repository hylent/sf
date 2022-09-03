package bs

import (
	"context"
	"github.com/hylent/sf/demo/proto"
	"github.com/hylent/sf/logger"
)

var (
	FooInstance = new(Foo)
)

type Foo struct {
	proto.UnimplementedFooServer
}

func (x *Foo) Get(ctx context.Context, in *proto.FooIn) (*proto.FooOut, error) {
	log.Debug("bs.Foo.Get", logger.M{
		"what": in.What,
	})

	if in.What == "wtf" {
		return nil, EFooInvalidWhat
	}

	out := new(proto.FooOut)
	out.What = in.What
	return out, nil
}
