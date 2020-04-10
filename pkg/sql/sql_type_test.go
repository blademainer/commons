package sql

import (
	"fmt"
	"gotest.tools/assert"
	"reflect"
	"testing"
)

type tagged struct {
	f      string `sql_type:"varchar(255)"`
	s      string `sql_type:"-"`
	cField string
}

func (t *tagged) SqlTypeCField() string {
	return "text"
}

type tagged2 struct {
	f      string `sql_type:"varchar(255)"`
	s      string `sql_type:"-"`
	cField string
}

func (t tagged2) SqlTypeCField() string {
	return "text"
}

func TestReflectParser_getFieldSqlType(t *testing.T) {
	tt := &tagged{}
	p := &parseContext{}
	tp := reflect.TypeOf(tt)
	p.fieldOfStruct = &tp
	of := tp.Elem()

	fField, _ := of.FieldByName("f")
	p.currentField = &fField
	sqlType, s := getFieldSqlType(p, fField)
	assert.Equal(t, sqlType, "varchar(255)")
	assert.Equal(t, s, true)

	sField, _ := of.FieldByName("s")
	p.currentField = &sField
	sqlType, s = getFieldSqlType(p, sField)
	assert.Equal(t, sqlType, "")
	assert.Equal(t, s, false)

	cField, _ := of.FieldByName("cField")
	method := reflect.TypeOf(tt).Method(0)
	fmt.Println(method)
	p.currentField = &cField
	sqlType, s = getFieldSqlType(p, cField)
	fmt.Println(sqlType)
	assert.Equal(t, sqlType, "text")
	assert.Equal(t, s, true)

	tt2 := tagged2{}
	p2 := &parseContext{}
	tp2 := reflect.TypeOf(tt2)
	p2.fieldOfStruct = &tp2

	cField2, _ := tp2.FieldByName("cField")
	method2 := reflect.TypeOf(tt2).Method(0)
	fmt.Println(method2)
	p2.currentField = &cField2
	sqlType, s = getFieldSqlType(p2, cField2)
	fmt.Println(sqlType)
	assert.Equal(t, sqlType, "text")
	assert.Equal(t, s, true)

}

func Test_buildFuncByFieldName(t *testing.T) {
	assert.Equal(t, buildFuncByFieldName(""), "")
	name := buildFuncByFieldName("hello")
	fmt.Println(name)
	assert.Equal(t, name, fmt.Sprintf("%s%s", getFuncWord, "Hello"))
}
