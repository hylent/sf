package db

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hylent/sf/logger"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
)

type AdapterMysql struct {
	Dsn string `yaml:"dsn"`

	db *sqlx.DB
}

func (x *AdapterMysql) Init() error {
	db, dbErr := sqlx.Open("mysql", x.Dsn)
	if dbErr != nil {
		return dbErr
	}

	if err := db.Ping(); err != nil {
		return err
	}

	x.db = db
	return nil
}

func (x *AdapterMysql) Execute(sql string, args ...any) (sql.Result, error) {
	startTp := time.Now()
	ret, err := x.db.Exec(sql, args...)
	log.Debug("sql_execute", logger.M{
		"sql":  sql,
		"args": fmt.Sprintf("%+v", args),
		"ret":  fmt.Sprintf("%#v", ret),
		"err":  fmt.Sprintf("%v", err),
		"cost": time.Now().Sub(startTp).Milliseconds(),
	})
	return ret, err
}

func (x *AdapterMysql) Select(out interface{}, sql string, args ...interface{}) error {
	startTp := time.Now()
	err := x.db.Select(out, sql, args...)
	log.Debug("sql_query", logger.M{
		"sql":  sql,
		"args": fmt.Sprintf("%+v", args),
		"err":  fmt.Sprintf("%v", err),
		"cost": time.Now().Sub(startTp).Milliseconds(),
	})
	return err
}

func (x *AdapterMysql) Row(out interface{}, sql string, args ...interface{}) error {
	startTp := time.Now()
	err := func() error {
		row := x.db.QueryRowx(sql, args...)
		if err := row.Err(); err != nil {
			return err
		}
		if err := row.StructScan(out); err != nil {
			return err
		}
		return nil
	}()
	log.Debug("sql_row", logger.M{
		"sql":  sql,
		"args": fmt.Sprintf("%+v", args),
		"err":  fmt.Sprintf("%v", err),
		"cost": time.Now().Sub(startTp).Milliseconds(),
	})
	return err
}

func (x *AdapterMysql) Get(out interface{}, sql string, args ...interface{}) error {
	startTp := time.Now()
	err := x.db.Get(out, sql, args...)
	log.Debug("sql_get", logger.M{
		"sql":  sql,
		"args": fmt.Sprintf("%+v", args),
		"err":  fmt.Sprintf("%v", err),
		"cost": time.Now().Sub(startTp).Milliseconds(),
	})
	return err
}

func (x *AdapterMysql) Insert(table string, data any) (int64, error) {
	inserts, dataErr := toMap(data)
	if dataErr != nil {
		return 0, dataErr
	}

	var fields []string
	var placeholders []string
	var params []any

	for f, v := range inserts {
		fields = append(fields, f)
		placeholders = append(placeholders, "?")
		params = append(params, v)
	}

	q := fmt.Sprintf(
		"insert into %s (%s) values (%s)",
		table,
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "),
	)

	ret, err := x.Execute(q, params...)
	if err != nil {
		return 0, err
	}

	return ret.LastInsertId()
}

func (x *AdapterMysql) Update(table string, data interface{}, where string, params ...any) error {
	updates, updatesErr := toMap(data)
	if updatesErr != nil {
		return updatesErr
	}

	// make where
	whereStr := ""
	if len(where) > 0 {
		whereStr = fmt.Sprintf(
			" where %s",
			where,
		)
	}

	var updateParts []string
	var newParams []any
	for k, v := range updates {
		updateParts = append(updateParts, fmt.Sprintf("%s=?", k))
		newParams = append(newParams, v)
	}
	newParams = append(newParams, params...)

	// make sql
	q := fmt.Sprintf(
		"update %s set %s%s",
		table,
		strings.Join(updateParts, ", "),
		whereStr,
	)

	_, err := x.Execute(q, newParams...)
	if err != nil {
		return err
	}

	return nil
}

func (x *AdapterMysql) Delete(table string, where string, params ...any) error {
	// make where
	whereStr := ""
	if len(where) > 0 {
		whereStr = fmt.Sprintf(
			" where %s",
			where,
		)
	}

	// make sql
	q := fmt.Sprintf(
		"delete from %s%s",
		table,
		whereStr,
	)

	_, err := x.Execute(q, params...)
	if err != nil {
		return err
	}

	return nil
}

type SelectOption struct {
	Order  string
	Limit  int64
	Offset int64
}

func (x *AdapterMysql) All(table string, out interface{}, opt *SelectOption, where string, params ...interface{}) error {
	// prepare fields
	fieldStr, fieldStrErr := getFieldStr(out)
	if fieldStrErr != nil {
		return fieldStrErr
	}

	// make where
	whereStr := ""
	if len(where) > 0 {
		whereStr = fmt.Sprintf(
			" where %s",
			where,
		)
	}

	// make order
	orderStr := ""
	if opt != nil && len(opt.Order) > 0 {
		orderStr = fmt.Sprintf(
			" order by %s",
			opt.Order,
		)
	}

	// make limit
	limitStr := ""
	if opt != nil && opt.Limit > 0 {
		if opt.Offset == 0 {
			limitStr = fmt.Sprintf(
				" limit %d",
				opt.Limit,
			)
		} else {
			limitStr = fmt.Sprintf(
				" limit %d, %d",
				opt.Offset,
				opt.Limit,
			)
		}
	}

	// make sql
	q := fmt.Sprintf(
		"select %s from %s%s%s%s",
		fieldStr,
		table,
		whereStr,
		orderStr,
		limitStr,
	)

	if err := x.Select(out, q, params...); err != nil {
		return err
	}

	return nil
}

func (x *AdapterMysql) Count(table string, where string, params ...interface{}) (int64, error) {
	// make where
	whereStr := ""
	if len(where) > 0 {
		whereStr = fmt.Sprintf(
			" where %s",
			where,
		)
	}

	// make sql
	q := fmt.Sprintf(
		"select count(*) from %s%s",
		table,
		whereStr,
	)

	var ret int64
	if err := x.Get(&ret, q, params...); err != nil {
		return 0, err
	}

	return ret, nil
}

func (x *AdapterMysql) Page(table string, out interface{}, order string, paginator *Paginator, where string, params ...interface{}) error {
	count, countErr := x.Count(table, where, params)
	if countErr != nil {
		return countErr
	}

	if paginator.check(count) {
		if err := x.All(
			table,
			out,
			&SelectOption{
				Order:  order,
				Limit:  paginator.Limit,
				Offset: paginator.Skip,
			},
			where,
			params...,
		); err != nil {
			return err
		}
	}

	return nil
}

func (x *AdapterMysql) First(table string, out interface{}, order string, where string, params ...interface{}) error {
	// prepare fields
	fieldStr, fieldStrErr := getFieldStr(out)
	if fieldStrErr != nil {
		return fieldStrErr
	}

	// make where
	whereStr := ""
	if len(where) > 0 {
		whereStr = fmt.Sprintf(
			" where %s",
			where,
		)
	}

	// make order
	orderStr := ""
	if len(order) > 0 {
		orderStr = fmt.Sprintf(
			" order by %s",
			order,
		)
	}

	// make sql
	q := fmt.Sprintf(
		"select %s from %s%s%s limit 1",
		fieldStr,
		table,
		whereStr,
		orderStr,
	)

	if err := x.Row(out, q, params...); err != nil {
		return err
	}

	return nil
}
