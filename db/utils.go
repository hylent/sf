package db

import (
	"fmt"
	"reflect"
	"strings"
)

func toMap(data interface{}) (map[string]interface{}, error) {
	if data == nil {
		return nil, nil
	}
	if duck, duckOk := data.(map[string]interface{}); duckOk {
		return duck, nil
	}
	rv := reflect.ValueOf(data)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("sql_invalid_data_kind: kind=%v", rv.Kind())
	}
	rt := rv.Type()
	ret := map[string]interface{}{}
	for i := 0; i < rt.NumField(); i++ {
		fRt := rt.Field(i)
		var name string
		if n := fRt.Tag.Get("db"); len(n) > 0 {
			name = n
		} else {
			name = strings.ToLower(fRt.Name)
		}
		fRv := rv.Field(i)
		if fRv.CanInterface() {
			ret[name] = fRv.Interface()
		}
	}
	return ret, nil
}

func getFieldStr(out interface{}) (string, error) {
	if out == nil {
		return "", fmt.Errorf("sql_nil_out")
	}
	if _, ok := out.(*map[string]interface{}); !ok {
		rv := reflect.ValueOf(out)
		if rv.Kind() != reflect.Ptr {
			return "", fmt.Errorf("sql_invalid_out_kind: kind=%v", rv.Kind())
		}
		elemRv := rv.Elem()
		if elemRv.Kind() != reflect.Struct {
			return "", fmt.Errorf("sql_invalid_out_elem_kind: kind=%v", elemRv.Kind())
		}
		var fields []string
		for i := 0; i < elemRv.NumField(); i++ {
			fieldRs := elemRv.Type().Field(i)
			field := fieldRs.Tag.Get("db")
			if len(field) < 1 {
				field = strings.ToLower(fieldRs.Name)
			}
			if strings.ContainsAny(field, " ") {
				fields = append(fields, field)
			} else {
				fields = append(fields, fmt.Sprintf("`%s`", field))
			}
		}
		if len(fields) > 0 {
			return strings.Join(fields, ", "), nil
		}
	}
	return "*", nil
}
