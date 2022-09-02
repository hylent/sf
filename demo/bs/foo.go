package bs

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/hylent/sf/demo/proto"
	"github.com/hylent/sf/logger"
	"github.com/hylent/sf/restful"
)

var (
	FooInstance = new(Foo)
)

type Foo struct {
	proto.UnimplementedFooServer
}

func (x *Foo) Get(ctx context.Context, in *proto.FooIn) (*proto.FooOut, error) {
	logger.Debug("bs.Foo.Get", logger.M{
		"what": in.What,
	})

	if in.What == "wtf" {
		return nil, EFooInvalidWhat
	}

	out := new(proto.FooOut)
	out.What = in.What
	return out, nil
}

func (x *Foo) HandleGet(ctx *gin.Context) {
	restful.HandleAsRestful(ctx, x.Get)
}
