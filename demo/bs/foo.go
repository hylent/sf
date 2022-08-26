package bs

import (
	"context"
	"github.com/hylent/sf/demo/api"
	"github.com/hylent/sf/logger"
)

var (
	FooImplInstance api.Foo = new(Foo)
)

type Foo struct {
}

func (x *Foo) Get(ctx context.Context, in *api.FooIn, out *api.FooOut) error {
	logger.Debug("bs.impl.Foo.Get", logger.M{
		"what": in.What,
	})

	if in.What == "fuck" {
		return api.EFooInvalidWhat
	}

	out.What = in.What
	return nil
}
