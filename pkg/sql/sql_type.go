package sql

import (
	"github.com/blademainer/commons/pkg/logger"
	"reflect"
	"strings"
)

const (
	tagSqlField = "sql_type"
	tagIgnore   = "-"
	getFuncWord = "SqlType"
)

// Get sqlType of the field.
// Returns true when the field is resolved
func getFieldSqlType(c *parseContext, field reflect.StructField) (string, bool) {
	lookup, ok := field.Tag.Lookup(tagSqlField)
	if ok {
		if lookup == tagIgnore {
			return "", false
		}
		tag, _ := parseTag(lookup)
		logger.Infof("get sqlType: %v of field: %v", tag, field.Name)
		return tag, true
	}
	funcName := buildFuncByFieldName(field.Name)
	t := *c.fieldOfStruct
	method, found := t.MethodByName(funcName)
	if !found {
		logger.Warnf("on type: %v not found tag: %v of field and not found method: %v", (*c.fieldOfStruct).Name(), tagSqlField, funcName)
		return "", false
	} else if method.Type.NumOut() != 1 {
		logger.Errorf("method: %v's out number is not 1", funcName)
		return "", false
	} else if method.Type.Out(0).Kind() != reflect.String {
		logger.Errorf("method: %v's out kind is not string", funcName)
		return "", false
	}
	//else if method.Type.NumIn() > 1 {
	//	builder := strings.Builder{}
	//	delimiter := ""
	//	for i := 0; i < method.Type.NumIn(); i++ {
	//		builder.WriteString(delimiter)
	//		builder.WriteString(fmt.Sprintf("%v", method.Type.In(i).String()))
	//		delimiter = ", "
	//	}
	//	logger.Errorf("method: %v's input args num is not zero, actual: %v args: %v", funcName, method.Type.NumIn(), builder.String())
	//	//return "", false
	//}
	logger.Warnf("found method: %v", method)

	in := make([]reflect.Value, 0)
	for i := 0; i < method.Type.NumIn(); i++ {
		mt := method.Type.In(i)
		value := reflect.New(mt)
		in = append(in, value.Elem())
	}

	call := method.Func.Call(in)
	if len(call) != 1 {
		logger.Errorf("")
		return "", false
	}
	value := call[0].String()
	logger.Infof("get sqlType: %v by func: %v", value, funcName)

	// get type by func
	return value, true
}

// gen the function of the field
func buildFuncByFieldName(fieldName string) string {
	if fieldName == "" {
		return ""
	}
	rs := make([]byte, len(fieldName)+len(getFuncWord))

	fieldArr := []byte(fieldName)
	// first word is upper
	h := strings.ToUpper(string(fieldName[0]))
	fieldArr[0] = []byte(h)[0]

	// append getFuncWord and fieldName
	index := 0
	for _, b := range []byte(getFuncWord) {
		rs[index] = b
		index++
	}
	for _, b := range fieldArr {
		rs[index] = b
		index++
	}

	return string(rs)
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
