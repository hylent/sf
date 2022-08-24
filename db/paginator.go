package db

const (
	PaginatorDefaultLimit = 10
)

type Paginator struct {
	Limit    int64
	Page     int64
	Skip     int64
	NumRows  int64
	NumPages int64
}

func (x *Paginator) check(numRows int64) bool {
	// validate
	if x.Limit < 1 {
		x.Limit = PaginatorDefaultLimit
	}
	if x.Page < 1 {
		x.Page = 1
	}
	x.Skip = x.Limit * (x.Page - 1)
	// set num rows
	if numRows < 1 {
		x.NumRows = 0
		x.NumPages = 0
	} else {
		x.NumRows = numRows
		x.NumPages = 1 + int64((numRows-1)/x.Limit)
	}
	// need query
	return x.Skip < x.NumRows
}
