package db

type Db interface {
	Insert(table string, data interface{}) (int64, error)
	Update(table string, data interface{}, where string, params ...interface{}) error
	Delete(table string, where string, params ...interface{}) error
	All(table string, out interface{}, order string, limit int, where string, params ...interface{}) error
	Page(table string, out interface{}, order string, paginator *Paginator, where string, params ...interface{}) error
	First(table string, out interface{}, order string, where string, params ...interface{}) error
	Aggregate(table string, out interface{}, where string, params ...interface{}) error
	Group(table string, out interface{}, by string, where string, params ...interface{}) error
}
