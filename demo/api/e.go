package api

import (
	"github.com/hylent/sf/restful"
)

const (
	EOk             restful.E = 0
	ENotImplemented restful.E = 1
	EInvalidParam   restful.E = 2

	EFooInvalidWhat restful.E = 12345
)
