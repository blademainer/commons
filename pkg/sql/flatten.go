package sql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/blademainer/commons/pkg/logger"
	"github.com/blademainer/commons/pkg/recoverable"
	"github.com/iancoleman/strcase"
	"reflect"
	"runtime"
)

type Parser interface {
	GetFields() (fields map[string]FieldEntry)
	GetValueMap(instance interface{}) (data map[string]interface{}, err error)
	ResolveFieldsFromMap(value map[string]interface{}, out interface{}) (err error)
}

type FieldType int

const (
	Node FieldType = iota
	Leaf
)

type FieldEntry struct {
	FieldType    FieldType
	GoFieldName  string
	TagFieldName string
	FullName     string
	Type         reflect.Type
	SqlType      string

	fieldValueFunc        FieldValueFunc
	resolveFieldValueFunc ResolveFieldValueFunc
	fieldPaths            []int
	parent                *FieldEntry
}

func IsComplexField(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Map, reflect.Array, reflect.Slice:
		return true
	default:
		return false
	}
}

func GetSqlType(t reflect.Type) string {
	if IsComplexField(t) {
		return "TEXT"
	}
	// see https://dev.mysql.com/doc/refman/8.0/en/integer-types.html
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return "INT"
	case reflect.Int64:
		return "BIGINT"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return "INT UNSIGNED"
	case reflect.Uint64:
		return "BIGINT UNSIGNED"
	case reflect.Bool:
		return "BOOLEAN"
	case reflect.String:
		return "VARCHAR(255)"
	default:
		return "TEXT"
	}
}

// Get value from field of instance
type FieldValueFunc func(instance interface{}) (fieldValue interface{}, err error)

// Setting value to the field of instance
type ResolveFieldValueFunc func(instance interface{}, fieldValue interface{}) (err error)
type FieldNameFunc func(field reflect.StructField) (string, error)

// Marshal func
type ComplexFieldMarshalFunc func(fieldValue interface{}) (string, error)
type ComplexFieldUnmarshalFunc func(raw string, ptr interface{}) error

var jsonFieldNameFunc FieldNameFunc = func(field reflect.StructField) (fieldName string, err error) {
	protoFieldName, found := field.Tag.Lookup("json")
	if protoFieldName == "-" {
		if logger.IsDebugEnabled() {
			logger.Debugf("field: %v is ignored", field.Name)
		}
		return "", fmt.Errorf("field: %v is ignored", field.Name)
	}

	tag, _ := parseTag(protoFieldName)
	if logger.IsDebugEnabled() {
		logger.Debugf("field: %v tag: %v parsedTag: %v", field.Name, protoFieldName, tag)
	}
	if tag == "" || !found {
		fieldName = strcase.ToSnake(field.Name)
	} else {
		fieldName = tag
	}
	return
}

var jsonComplexFieldMarshalFunc ComplexFieldMarshalFunc = func(fieldValue interface{}) (string, error) {
	bf := bytes.NewBuffer([]byte{})
	jsonEncoder := json.NewEncoder(bf)
	jsonEncoder.SetEscapeHTML(false)
	err := jsonEncoder.Encode(fieldValue)
	//fmt.Println("第二种解决办法：", bf.String())
	//value, err := json.Marshal(fieldValue)
	if err != nil {
		return "", err
	}

	return bf.String(), nil
}

var jsonComplexFieldUnmarshalFunc ComplexFieldUnmarshalFunc = func(raw string, ptr interface{}) error {
	err := json.Unmarshal([]byte(raw), ptr)
	if err != nil {
		return err
	}

	return nil
}

type reflectParser struct {
	instance                  interface{}
	t                         reflect.Type
	entries                   map[string]FieldEntry
	fieldNameFunc             FieldNameFunc
	complexFieldMarshalFunc   ComplexFieldMarshalFunc
	complexFieldUnmarshalFunc ComplexFieldUnmarshalFunc
}

type parseContext struct {
	currentField  *reflect.StructField
	fieldOfStruct *reflect.Type
}

func (p *reflectParser) ResolveFieldsFromMap(value map[string]interface{}, out interface{}) (err error) {
	if err := checkType(p.t, out); err != nil {
		return err
	}

	of := reflect.TypeOf(out)
	if of.Kind() != reflect.Ptr {
		err = fmt.Errorf("type: %v is not ptr", of)
		return err
	}
	if err = checkType(p.t, out); err != nil {
		return err
	}

	for k, v := range value {
		entry, exists := p.entries[k]
		if !exists {
			logger.Warnf("unknown field: %v", k)
			continue
		}

		err := entry.resolveFieldValueFunc(out, v)
		if err != nil {
			if IsNilError(err) {
				continue
			}
			return err
		}
	}

	return nil
}

func (p *reflectParser) GetValueMap(instance interface{}) (data map[string]interface{}, err error) {
	if err := checkType(p.t, instance); err != nil {
		return nil, err
	}

	data = make(map[string]interface{})
	fields := p.GetFields()
	for field, entry := range fields {
		if entry.FieldType == Node {
			continue
		}
		if entry.fieldValueFunc == nil {
			err = fmt.Errorf("FieldValueFunc of entry: %v is null", entry)
			return
		}
		fieldValue, err := entry.fieldValueFunc(instance)
		if err != nil {
			if IsNilError(err) {
				continue
			}
			return nil, err
		}
		data[field] = fieldValue
	}

	return
}

// flatten multilayer obj to key/value map
func NewParser(demo interface{}) (Parser, error) {
	p := &reflectParser{}
	p.fieldNameFunc = jsonFieldNameFunc
	p.complexFieldMarshalFunc = jsonComplexFieldMarshalFunc
	p.complexFieldUnmarshalFunc = jsonComplexFieldUnmarshalFunc

	t := reflect.TypeOf(demo)
	p.t = t

	parseContext := &parseContext{}
	m, err := p.convert(parseContext)
	if err != nil {
		return nil, err
	}
	p.entries = m
	return p, nil
}

func (p *reflectParser) GetFields() (fields map[string]FieldEntry) {
	return p.entries
}

func (p *reflectParser) convert(c *parseContext) (fields map[string]FieldEntry, err error) {
	t := p.t
	for t.Kind() == reflect.Ptr {
		logger.Debugf("%v is pointer", t)
		t = t.Elem()
	}
	fieldPaths := make([]int, 0)
	c.fieldOfStruct = &t
	return p.convertStruct(c, fieldPaths, t)
}

func (p *reflectParser) convertStruct(c *parseContext, fieldPaths []int, t reflect.Type) (fields map[string]FieldEntry, err error) {
	fields = make(map[string]FieldEntry)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		c.currentField = &field

		fieldPathsTmp := append(fieldPaths, i)
		fieldsResult, err := p.parseField(c, fieldPathsTmp, field)
		if logger.IsDebugEnabled() {
			logger.Debugf("parse field: %v to entries: %v", field, fieldsResult)
		}
		if fieldsResult == nil || err != nil {
			continue
		}

		for k, v := range fieldsResult {
			fields[k] = v
		}
	}

	return
}

func (p *reflectParser) findFieldOfInstance(field FieldEntry, instance interface{}) (*reflect.Value, error) {
	//defer func() {
	//	err := recover()
	//	fmt.Println(err)
	//}()

	t := p.t
	of := reflect.ValueOf(instance)

	for _, path := range field.fieldPaths {

		for t.Kind() == reflect.Ptr {
			if logger.IsDebugEnabled() {
				logger.Debugf("%v is pointer", t)
			}
			t = t.Elem()
		}

		if err := checkNil(of); err != nil {
			return nil, err
		}
		for of.Kind() == reflect.Ptr {
			of = of.Elem()
			if err := checkNil(of); err != nil {
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
		if err := checkNil(of); err != nil {
			//TODO just support first layer of pointer, need support multiply layer of pointer?
			logger.Infof("of.Type: %v of.Type().Elem(): %v", of.Type(), of.Type().Elem())
			if IsComplexField(of.Type()) {
				of.Set(reflect.New(of.Type()).Elem())
			} else {
				switch of.Kind() {
				case reflect.Ptr:
					of.Set(reflect.New(of.Type().Elem()))
				//case reflect.Map, reflect.Array, reflect.Slice:
				//	of.Set(reflect.New(of.Type()).Elem())
				default:
					of.Set(reflect.New(of.Type()))
				}
			}

		}
		if logger.IsDebugEnabled() {
			logger.Debugf("after instance: %v type: %v", instance, reflect.TypeOf(instance))

			logger.Debugf("t type: %v fields: %v path: %v", t, t.NumField(), path)
		}
		structField := t.Field(path)
		t = structField.Type
	}
	return &of, nil
}

func (p *reflectParser) findField(field FieldEntry, instance interface{}) (*reflect.Value, error) {
	t := p.t
	of := reflect.ValueOf(instance)

	for _, path := range field.fieldPaths {
		for t.Kind() == reflect.Ptr {
			if logger.IsDebugEnabled() {
				logger.Debugf("%v is pointer", t)
			}
			t = t.Elem()
		}
		if err := checkNil(of); err != nil {
			if logger.IsDebugEnabled() {
				logger.Debugf("failed to find field: %v, error: %v", field.FullName, err.Error())
			}
			return nil, err
		}
		for of.Kind() == reflect.Ptr {
			of = of.Elem()
			if err := checkNil(of); err != nil {
				if logger.IsDebugEnabled() {
					logger.Debugf("failed to find ptr field: %v, error: %v", field.FullName, err.Error())
				}

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

func (p *reflectParser) getValue(of reflect.Value) (interface{}, error) {
	//var value interface{}
	//
	if IsComplexField(of.Type()) {
		value, err := p.complexFieldMarshalFunc(of.Interface())
		if err != nil {
			return nil, err
		}
		return value, nil
	}
	return of.Interface(), nil
}

func (p *reflectParser) resolveFieldValue(field *reflect.Value, fieldValue interface{}) (err error) {
	//var value interface{}
	//
	defer recoverable.RecoverWithHandle(func(i interface{}) error {
		runtime.Caller(1)
		logger.Errorf("resolveField: %v value: %v with error: %v", field.String(), fieldValue, i)
		err = fmt.Errorf("resolveField: %v value: %v with error: %v", field.String(), fieldValue, i)
		return err
	})

	if IsComplexField(field.Type()) {
		// convert array to string
		v := reflect.ValueOf(fieldValue)

		if v.Kind() != reflect.String {
			switch v.Kind() {
			case reflect.Slice, reflect.Array:
				//switch v.Elem().Kind() {
				//// uint array to string
				//case reflect.Uint8, reflect.Int8:
				s := string(v.Interface().([]byte))
				v = reflect.ValueOf(s)
				//}
			}

			//return fmt.Errorf("field type: %v is map or array, but value type is not string. actual kind: %v", field.Type(), v.Kind())
		}

		vt := v.Type()
		switch vt.Kind() {
		case reflect.Slice, reflect.Array:
			switch vt.Elem().Kind() {
			// uint array to string
			case reflect.Uint8, reflect.Int8:
				s := string(v.Interface().([]byte))
				v = reflect.ValueOf(s)
			}
		}

		if logger.IsDebugEnabled() {
			logger.Debugf("instance unmarshal before: %v", field)
		}
		// get addr of the map value
		err := p.complexFieldUnmarshalFunc(v.String(), field.Addr().Interface())
		if err != nil {
			logger.Errorf("failed unmarshal field: %v from: %v error: %v", field, fieldValue, err.Error())
			return err
		}
		if logger.IsDebugEnabled() {
			logger.Debugf("instance unmarshal after: %v from: %v", field, fieldValue)
		}
		return nil
	}
	v := reflect.ValueOf(fieldValue)
	vt := reflect.TypeOf(fieldValue)
	switch field.Type().Kind() {
	case reflect.Int32:
		v = reflect.ValueOf(int32(v.Int()))
	case reflect.Uint32:
		v = reflect.ValueOf(uint32(v.Uint()))
	case reflect.Float32:
		v = reflect.ValueOf(float32(v.Float()))
	case reflect.Bool:
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			v = reflect.ValueOf(v.Int() > 0)
		}
	case reflect.String:
		switch vt.Kind() {
		case reflect.Slice, reflect.Array:
			switch vt.Elem().Kind() {
			// uint array to string
			case reflect.Uint8, reflect.Int8:
				s := string(v.Interface().([]byte))
				v = reflect.ValueOf(s)
			}
		}
	}

	field.Set(v)
	return nil
}

func checkType(t reflect.Type, instance interface{}) error {
	it := reflect.TypeOf(instance)
	if t != it {
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		for it.Kind() == reflect.Ptr {
			it = it.Elem()
		}
		if t != it {
			err := fmt.Errorf("expect type: %v actual instance type: %v ", t, reflect.TypeOf(instance))
			return err
		}
	}
	return nil
}

func (p *reflectParser) generateFieldValueFunc(field FieldEntry) FieldValueFunc {
	return func(instance interface{}) (interface{}, error) {

		if logger.IsDebugEnabled() {
			logger.Debugf("Finding field: %v value on instance: %v", field, instance)
		}

		findField, err := p.findField(field, instance)
		if err != nil {
			if logger.IsDebugEnabled() {
				logger.Debugf("failed to find field: %v, error: %v", field.FullName, err.Error())
			}
			return nil, err
		}

		of := *findField

		if err := checkNil(of); err != nil {
			if logger.IsDebugEnabled() {
				logger.Debugf("check nil with error: %v", err.Error())
			}
			return nil, nil
		}

		value, err := p.getValue(of)
		if err != nil {
			logger.Errorf("failed to get value of field: %v error: %v", field.FullName, err.Error())
			return nil, err
		}
		if value == nil {
			if logger.IsDebugEnabled() {
				logger.Debugf("failed to get value of field: %v", field)
			}
		}

		return value, nil
	}
}

func (p *reflectParser) generateResolveFieldValueFunc(field FieldEntry) ResolveFieldValueFunc {
	return func(instance interface{}, fieldValue interface{}) error {
		if fieldValue == nil {
			return nil
		}

		if logger.IsDebugEnabled() {
			logger.Debugf("Finding field: %v value on instance: %v", field, instance)
		}

		for parent := field.parent; parent != nil; parent = field.parent {
			findField, err := p.findFieldOfInstance(field, instance)
			if err != nil {
				return err
			}
			if err := checkNil(*findField); err != nil {
				if logger.IsDebugEnabled() {
					logger.Debugf("---> check nil with error: %v", err.Error())
				}
				continue
			}

			var value reflect.Value
			if parent.Type.Kind() == reflect.Ptr {
				value = reflect.New(parent.Type.Elem())
			} else {
				value = reflect.New(parent.Type)
			}

			findField.Set(value.Elem())
			if logger.IsDebugEnabled() {
				logger.Debugf("---> setting parent value: %v to field: %v canset: %v", value, findField, findField.CanSet())
			}
		}

		findField, err := p.findFieldOfInstance(field, instance)
		if err != nil {
			return err
		}

		//value := reflect.New(findField.Type())
		//value.Set(reflect.ValueOf(fieldValue))
		err = p.resolveFieldValue(findField, fieldValue)
		//findField.Set(reflect.ValueOf(fieldValue))
		if logger.IsDebugEnabled() {
			logger.Debugf("==> setting value: %v to field: %v canset: %v", fieldValue, findField.Type(), findField.CanSet())
		}

		return err
	}
}

type NilError struct {
	error string
}

func (n *NilError) Error() string {
	return n.error
}

func IsNilError(err error) bool {
	if err == nil {
		return false
	}
	_, castOk := (err).(*NilError)
	return castOk
}

func checkNil(of reflect.Value) *NilError {
	switch of.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		if of.IsNil() {
			err := fmt.Sprintf("field type: %v is nil", of)
			//panic(err)
			return &NilError{err}
		}
	}
	return nil
}

func (p *reflectParser) parseField(c *parseContext, fieldPaths []int, field reflect.StructField) (entries map[string]FieldEntry, err error) {
	entries = make(map[string]FieldEntry)

	fieldName, err := p.fieldNameFunc(field)
	if err != nil {
		return nil, err
	}

	t := field.Type
	if logger.IsDebugEnabled() {
		logger.Debugf("field.Type: %v", t)
	}
	for t.Kind() == reflect.Ptr {
		if logger.IsDebugEnabled() {
			logger.Debugf("%v is pointer", t)
		}
		t = t.Elem()
	}
	kind := t.Kind()
	switch kind {
	case reflect.Struct:
		c.fieldOfStruct = &field.Type
		convertStruct, err := p.convertStruct(c, fieldPaths, t)
		fe := FieldEntry{
			GoFieldName:  field.Name,
			FullName:     fieldName,
			TagFieldName: fieldName,
			Type:         t,
			fieldPaths:   fieldPaths,
			FieldType:    Node,
		}
		entries[fieldName] = fe
		if err != nil || len(convertStruct) == 0 {
			if logger.IsDebugEnabled() {
				logger.Debugf("type: %v has no fields result. error: %v", t, err)
			}
		} else {
			for k, v := range convertStruct {
				key := fmt.Sprintf("%v_%v", fieldName, k)
				v.FullName = key
				v.parent = &fe
				entries[key] = v
			}
		}
	default:
		fe := FieldEntry{
			GoFieldName:  field.Name,
			FullName:     fieldName,
			TagFieldName: fieldName,
			Type:         t,
			fieldPaths:   fieldPaths,
			FieldType:    Leaf,
		}

		f := p.generateFieldValueFunc(fe)
		fe.fieldValueFunc = f

		resolveFieldValueFunc := p.generateResolveFieldValueFunc(fe)
		if resolveFieldValueFunc == nil {
			err = fmt.Errorf("resolveFieldValueFunc of field: %v is null", fe)
		}
		fe.resolveFieldValueFunc = resolveFieldValueFunc

		fe.SqlType = p.GetSqlType(c, field)

		entries[fieldName] = fe
	}

	return entries, nil
}

func (p *reflectParser) GetSqlType(c *parseContext, fe reflect.StructField) string {
	tag, resolved := getFieldSqlType(c, fe)
	if !resolved {
		return GetSqlType(fe.Type)
	}
	return tag
}
