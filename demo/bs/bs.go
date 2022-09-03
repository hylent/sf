package bs

import (
	"github.com/hylent/sf/logger"
	"github.com/hylent/sf/server"
)

var log = logger.NewLogger(nil, "demo/bs")

const (
	EOk             server.E = 0
	ENotImplemented server.E = 1
	EInvalidParam   server.E = 2

	EFooInvalidWhat server.E = 12345
)
