package api

import "context"

type Foo interface {
	Get(ctx context.Context, in *FooIn, out *FooOut) error
}

type FooIn struct {
	What string `json:"what" form:"what" pb:"what"`
}

type FooOut struct {
	What string `json:"what"`
}
