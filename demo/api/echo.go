package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Echo struct {
}

func (x *Echo) HandleGet(ctx *gin.Context) {
	ctx.String(http.StatusOK, "OK")
}
