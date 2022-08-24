package api

import (
	"github.com/gin-gonic/gin"
	"github.com/hylent/sf/demo/bs"
	"github.com/hylent/sf/restful"
)

type Exp struct {
}

func (x *Exp) HandleGet(ctx *gin.Context) {
	restful.Handle(ctx, bs.ExpInstance.Get)
}

func (x *Exp) HandlePost(ctx *gin.Context) {
	restful.Handle(ctx, bs.ExpInstance.Post)
}
