package sql

import (
	"fmt"
	"github.com/blademainer/commons/pkg/logger"
	"github.com/iancoleman/strcase"
	"reflect"
	"strings"
)

type Parser struct {
	instance interface{}
	t        reflect.Type
	data     map[string]FieldType
}

type FieldType struct {
	GoFieldName    string
	TagFieldName   string
	SqlFieldName   string
	Kind           reflect.Kind
	Path           []int
	FieldValueFunc FieldValueFunc
}

type FieldValueFunc func(instance interface{}) (interface{}, error)

// flatten multilayer obj to key/value map
func NewParser(demo interface{}) *Parser {
	p := &Parser{
		instance: demo,
	}
	return p
}

func (p *Parser) Parse() (fields map[string]FieldType, err error) {
	return p.convertJson(p.instance)
}

func (p *Parser) convertJson(demo interface{}) (fields map[string]FieldType, err error) {
	t := reflect.TypeOf(demo)
	p.t = t
	for t.Kind() == reflect.Ptr {
		logger.Debugf("%v is pointer", t)
		t = t.Elem()
	}
	fieldPaths := make([]int, 0)
	return p.convertStruct(fieldPaths, t)
}

func (p *Parser) convertStruct(fieldPaths []int, t reflect.Type) (fields map[string]FieldType, err error) {
	fields = make(map[string]FieldType)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		fieldPathsTmp := append(fieldPaths, i)
		fieldsResult, err := p.ParseField(fieldPathsTmp, field)
		if fieldsResult == nil || err != nil {
			continue
		}

		for k, v := range fieldsResult {
			fields[k] = v
		}
	}

	return
}

func (p *Parser) findField(field FieldType, instance interface{}) (*reflect.Value, error) {
	t := p.t
	of := reflect.ValueOf(instance)

	for _, path := range field.Path {
		for t.Kind() == reflect.Ptr {
			if logger.IsDebugEnabled() {
				logger.Debugf("%v is pointer", t)
			}
			t = t.Elem()
		}

		if err := CheckNil(of); err != nil {
			return nil, err
		}
		for of.Kind() == reflect.Ptr {
			of = of.Elem()
			if err := CheckNil(of); err != nil {
				return nil, err
			}
		}
		if logger.IsDebugEnabled() {
			logger.Debugf("===========================")
			logger.Debugf("field: %v ", field)
			logger.Debugf("before instance: %v", instance)
			logger.Debugf("instance type: %v numField: %v path: %v", reflect.TypeOf(instance), of.NumField(), path)
		}
		of = of.Field(path)
		if logger.IsDebugEnabled() {
			logger.Debugf("after instance: %v type: %v", instance, reflect.TypeOf(instance))

			logger.Debugf("t type: %v fields: %v path: %v", t, t.NumField(), path)
		}
		structField := t.Field(path)
		t = structField.Type
	}
	return &of, nil
}

func getValue(of reflect.Value) interface{} {
	//var value interface{}
	//
	//switch of.Kind() {
	////case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
	////	value = of.Int()
	////case reflect.String:
	////	value = of.String()
	//default:
	//	return of.Interface()
	//}
	return of.Interface()
}

func (p *Parser) GetValueFunc(field FieldType) FieldValueFunc {
	return func(instance interface{}) (interface{}, error) {
		if reflect.TypeOf(instance) != p.t {
			err := fmt.Errorf("expect type: %v actual instance type: %v ", p.t, reflect.TypeOf(instance))
			return nil, err
		}

		if logger.IsDebugEnabled() {
			logger.Debugf("Finding field: %v value on instance: %v", field, instance)
		}

		findField, err := p.findField(field, instance)
		if err != nil {
			return nil, err
		}

		of := *findField

		//for of.Kind() == reflect.Ptr {
		//	of = of.Elem()
		//}

		//fmt.Println("of === ", reflect.ValueOf(of))

		if err := CheckNil(of); err != nil {
			if logger.IsDebugEnabled() {
				logger.Debugf("check nil with error: %v", err.Error())
			}
			return nil, nil
		}

		value := getValue(of)
		if value == nil {
			if logger.IsDebugEnabled() {
				logger.Debugf("failed to get value of field: %v", field)
			}
		}

		return value, nil
	}
}

func CheckNil(of reflect.Value) error {
	switch of.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		if of.IsNil() {
			err := fmt.Errorf("field type: %v is nil", of)
			return err
		}
	}
	return nil
}

func (p *Parser) ParseField(fieldPaths []int, field reflect.StructField) (fieldType map[string]FieldType, err error) {
	protoFieldName, found := field.Tag.Lookup("json")
	if protoFieldName == "-" {
		if logger.IsDebugEnabled() {
			logger.Debugf("field: %v is ignored\n", field.Name)
		}
		return nil, nil
	}

	fieldType = make(map[string]FieldType)
	tag, _ := parseTag(protoFieldName)
	if logger.IsDebugEnabled() {
		logger.Debugf("field: %v tag: %v parsedTag: %v\n", field.Name, protoFieldName, tag)
	}
	var fieldName string
	if tag == "" || !found {
		fieldName = strcase.ToSnake(field.Name)
	} else {
		fieldName = tag
	}

	if logger.IsDebugEnabled() {
		logger.Debugf("field.Type: %v", field.Type)
	}
	t := field.Type
	for t.Kind() == reflect.Ptr {
		if logger.IsDebugEnabled() {
			logger.Debugf("%v is pointer", t)
		}
		t = t.Elem()
	}
	kind := t.Kind()
	switch kind {
	case reflect.Struct:
		convertStruct, err := p.convertStruct(fieldPaths, t)
		if err != nil || len(convertStruct) == 0 {
			if logger.IsDebugEnabled() {
				logger.Debugf("type: %v has no fields result. error: %v", t, err)
			}
		} else {
			for k, v := range convertStruct {
				key := fmt.Sprintf("%v_%v", fieldName, k)
				v.SqlFieldName = key
				fieldType[key] = v
			}
		}
	default:
		ft := FieldType{
			GoFieldName:  field.Name,
			SqlFieldName: fieldName,
			TagFieldName: fieldName,
			Kind:         kind,
			Path:         fieldPaths,
		}

		f := p.GetValueFunc(ft)
		ft.FieldValueFunc = f

		fieldType[fieldName] = ft
	}

	return fieldType, nil
}

// tagOptions is the string following a comma in a struct field's "json"
// tag, or the empty string. It does not include the leading comma.
type tagOptions string

// parseTag splits a struct field's json tag into its name and
// comma-separated options.
func parseTag(tag string) (string, tagOptions) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tagOptions(tag[idx+1:])
	}
	return tag, tagOptions("")
}

// Contains reports whether a comma-separated list of options
// contains a particular substr flag. substr must be surrounded by a
// string boundary or commas.
func (o tagOptions) Contains(optionName string) bool {
	if len(o) == 0 {
		return false
	}
	s := string(o)
	for s != "" {
		var next string
		i := strings.Index(s, ",")
		if i >= 0 {
			s, next = s[:i], s[i+1:]
		}
		if s == optionName {
			return true
		}
		s = next
	}
	return false
}
