package reiner

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"unicode"
)

// https://github.com/bgaifullin/dbr
// https://github.com/bgaifullin/dbr/blob/master/load.go
var (
	ErrNotFound           = errors.New("Reiner: not found")
	ErrNotSupported       = errors.New("Reiner: not supported")
	ErrTableNotSpecified  = errors.New("Reiner: table not specified")
	ErrColumnNotSpecified = errors.New("Reiner: column not specified")
	ErrInvalidPointer     = errors.New("Reiner: attempt to load into an invalid pointer")
	ErrPlaceholderCount   = errors.New("Reiner: wrong placeholder count")
	ErrInvalidSliceLength = errors.New("Reiner: length of slice is 0. length must be >= 1")
	ErrCantConvertToTime  = errors.New("Reiner: can't convert to time.Time")
	ErrInvalidTimestring  = errors.New("Reiner: invalid time string")
)

// Load loads any value from sql.Rows
func load(rows *sql.Rows, value interface{}) (int, error) {
	defer rows.Close()
	if value == nil {
		count := 0
		for rows.Next() {
			count++
		}
		return count, nil
	}

	column, err := rows.Columns()
	if err != nil {
		return 0, err
	}

	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return 0, ErrInvalidPointer
	}

	isSlice := v.Elem().Kind() == reflect.Slice && v.Elem().Type().Elem().Kind() != reflect.Uint8
	if isSlice {
		v.Elem().Set(reflect.MakeSlice(v.Type().Elem(), 0, v.Elem().Cap()))
	}

	v = v.Elem()

	count := 0
	var elemType reflect.Type
	if isSlice {
		elemType = v.Type().Elem()
	} else {
		elemType = v.Type()
	}
	extractor, err := findExtractor(elemType)
	if err != nil {
		return count, err
	}
	for rows.Next() {
		var elem reflect.Value
		if isSlice {
			elem = reflect.New(v.Type().Elem()).Elem()
		} else {
			elem = v
		}
		ptr := extractor(column, elem)

		err = rows.Scan(ptr...)
		if err != nil {
			return count, err
		}
		count++
		if isSlice {
			v.Set(reflect.Append(v, elem))
		} else {
			break
		}
	}
	return count, nil
}

type dummyScanner struct{}

func (dummyScanner) Scan(interface{}) error {
	return nil
}

type keyValueMap map[string]interface{}

type kvScanner struct {
	column string
	m      keyValueMap
}

func (kv *kvScanner) Scan(v interface{}) error {
	kv.m[kv.column] = v
	return nil
}

type pointersExtractor func(columns []string, value reflect.Value) []interface{}

var (
	dummyDest       sql.Scanner = dummyScanner{}
	typeScanner                 = reflect.TypeOf((*sql.Scanner)(nil)).Elem()
	typeKeyValueMap             = reflect.TypeOf(keyValueMap(nil))
)

// structMap builds index to fast lookup fields in struct
func structMap(t reflect.Type) map[string][]int {
	m := make(map[string][]int)
	structTraverse(m, t, nil)
	return m
}

var (
	typeValuer = reflect.TypeOf((*driver.Valuer)(nil)).Elem()
)

func camelCaseToSnakeCase(name string) string {
	buf := new(bytes.Buffer)

	runes := []rune(name)

	for i := 0; i < len(runes); i++ {
		buf.WriteRune(unicode.ToLower(runes[i]))
		if i != len(runes)-1 && unicode.IsUpper(runes[i+1]) &&
			(unicode.IsLower(runes[i]) || unicode.IsDigit(runes[i]) ||
				(i != len(runes)-2 && unicode.IsLower(runes[i+2]))) {
			buf.WriteRune('_')
		}
	}

	return buf.String()

}

func structTraverse(m map[string][]int, t reflect.Type, head []int) {
	if t.Implements(typeValuer) {
		return
	}
	switch t.Kind() {
	case reflect.Ptr:
		structTraverse(m, t.Elem(), head)
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath != "" && !field.Anonymous {
				// unexported
				continue
			}
			tag := field.Tag.Get("db")
			if tag == "-" {
				// ignore
				continue
			}
			if tag == "" {
				// no tag, but we can record the field name
				//tag = camelCaseToSnakeCase(field.Name)
				tag = field.Name
			}
			if _, ok := m[tag]; !ok {
				m[tag] = append(head, i)
			}
			structTraverse(m, field.Type, append(head, i))
		}
	}
}

func getStructFieldsExtractor(t reflect.Type) pointersExtractor {
	mapping := structMap(t)
	return func(columns []string, value reflect.Value) []interface{} {
		var ptr []interface{}
		for _, key := range columns {
			if index, ok := mapping[key]; ok {
				ptr = append(ptr, value.FieldByIndex(index).Addr().Interface())
			} else {
				ptr = append(ptr, dummyDest)
			}
		}
		return ptr
	}
}

func getIndirectExtractor(extractor pointersExtractor) pointersExtractor {
	return func(columns []string, value reflect.Value) []interface{} {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		return extractor(columns, value.Elem())
	}
}

func mapExtractor(columns []string, value reflect.Value) []interface{} {
	if value.IsNil() {
		value.Set(reflect.MakeMap(value.Type()))
	}
	m := value.Convert(typeKeyValueMap).Interface().(keyValueMap)
	var ptr []interface{}
	for _, c := range columns {
		ptr = append(ptr, &kvScanner{column: c, m: m})
	}
	return ptr
}

func dummyExtractor(columns []string, value reflect.Value) []interface{} {
	return []interface{}{value.Addr().Interface()}
}

func findExtractor(t reflect.Type) (pointersExtractor, error) {
	if reflect.PtrTo(t).Implements(typeScanner) {
		return dummyExtractor, nil
	}

	switch t.Kind() {
	case reflect.Map:
		if !t.ConvertibleTo(typeKeyValueMap) {
			return nil, fmt.Errorf("expected %v, got %v", typeKeyValueMap, t)
		}
		return mapExtractor, nil
	case reflect.Ptr:
		inner, err := findExtractor(t.Elem())
		if err != nil {
			return nil, err
		}
		return getIndirectExtractor(inner), nil
	case reflect.Struct:
		return getStructFieldsExtractor(t), nil
	}
	return dummyExtractor, nil
}
