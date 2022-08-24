package bs

import "fmt"

type E int

func (x E) Error() string {
	return fmt.Sprintf("E%06d", int(x))
}

func (x E) Code() int {
	return int(x)
}

const (
	EOk             E = 0
	ENotImplemented E = 1
	EInvalidParam   E = 2
)
