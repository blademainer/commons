package util

import (
	"encoding/json"
	"fmt"
	"github.com/blademainer/commons/pkg/logger"
	"github.com/iancoleman/strcase"
	"reflect"
	"strings"
)

type FieldDoc struct {
	Name     string `json:"name"`
	Desc     string `json:"desc"`
	Required string `json:"required"`
	Remark   string `json:"remark"`
	Type     string `json:"type"`
	Doc      *Doc   `json:"doc"`
}

type Doc struct {
	Fields   []FieldDoc `json:"fields"`
	DemoJson string     `json:"demo_json"`
}

type Parser struct {
	Demo interface{}
	T    reflect.Type
}

func ParseDoc(config interface{}) (*Doc, error) {
	p := &Parser{}
	p.Demo = config
	t := reflect.TypeOf(config)
	p.T = t

	marshal, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	fields, err := p.ParseType(t)
	if err != nil {
		logger.Errorf("failed to parse type: %v error: %v", t, err.Error())
	}

	doc := &Doc{}
	doc.Fields = fields
	doc.DemoJson = string(marshal)
	return doc, nil
}

func (p *Parser) ParseType(t reflect.Type) ([]FieldDoc, error) {

	for t.Kind() == reflect.Ptr {
		logger.Debugf("%v is pointer", t)
		t = t.Elem()
	}

	fields := make([]FieldDoc, 0)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		fieldsResult, err := p.parseField(field)
		if logger.IsDebugEnabled() {
			logger.Debugf("parse field: %v to entries: %v", field, fieldsResult)
		}
		if fieldsResult == nil || err != nil {
			continue
		}

		for _, fieldDoc := range fieldsResult {
			fields = append(fields, fieldDoc)
		}
	}
	return fields, nil
}

func (p *Parser) parseField(field reflect.StructField) (entries []FieldDoc, err error) {
	entries = make([]FieldDoc, 0)
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
		entry, err := p.parseFieldDoc(field)
		if err != nil {
			logger.Errorf("failed to parse field: %v in struct: %v error: %v", field.Name, field, err.Error())
			return nil, err
		}

		of := reflect.ValueOf(p.Demo)

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

		fv := of.FieldByName(field.Name)
		vv := fv.Interface()
		subDoc, err := ParseDoc(vv)
		if err != nil {
			logger.Errorf("failed to parse field: %v struct: %v, error: %v", field.Name, t, err.Error())
			return nil, err
		}
		entry.Doc = subDoc
		entries = append(entries, *entry)
	default:
		entry, err := p.parseFieldDoc(field)
		if err != nil {
			logger.Errorf("failed to parse field: %v in struct: %v error: %v", field.Name, field, err.Error())
			return nil, err
		}
		entries = append(entries, *entry)
	}

	return entries, nil
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

func (p *Parser) parseFieldDoc(field reflect.StructField) (doc *FieldDoc, err error) {
	protoFieldName, found := field.Tag.Lookup("json")
	if protoFieldName == "-" {
		if logger.IsDebugEnabled() {
			logger.Debugf("field: %v is ignored", field.Name)
		}
		return nil, fmt.Errorf("field: %v is ignored", field.Name)
	}

	tag, _ := parseTag(protoFieldName)
	if logger.IsDebugEnabled() {
		logger.Debugf("field: %v tag: %v parsedTag: %v", field.Name, protoFieldName, tag)
	}
	doc = &FieldDoc{}
	if tag == "" || !found {
		doc.Name = strcase.ToSnake(field.Name)
	} else {
		doc.Name = tag
	}
	doc.Type = field.Type.Kind().String()

	docTag, ok := field.Tag.Lookup("doc")
	if !ok {
		return doc, nil
	}

	m := make(map[string]interface{})
	// split by comma
	pairStrings := strings.Split(docTag, ",")
	for _, pairString := range pairStrings {
		kv := strings.Split(pairString, "=")
		if len(kv) != 2 {
			err = fmt.Errorf("tag: %v of field: %v is illegal", docTag, field.Name)
			return nil, err
		}
		m[kv[0]] = kv[1]
	}
	marshal, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	logger.Infof("found doc: %v", string(marshal))
	err = json.Unmarshal(marshal, doc)
	return
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
