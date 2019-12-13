package sql

import (
	"fmt"
	"github.com/blademainer/commons/pkg/logger"
	"gotest.tools/assert"
	"testing"
)

type A struct {
	B B
	C C
}

type B struct {
	C C
}

type C struct {
	Name     string
	Age      int
	Adult    bool
	Nullable *string
}

func ExampleParser_Parse() {
	logger.SetLevel("debug")
	name := "test"
	p := &A{
		B: B{
			C: C{
				Name: name,
			},
		},
	}
	parser := NewParser(p)
	proto, err := parser.Parse()
	fmt.Printf("fields: %v err: %v\n", proto, err)

	fieldType := proto["b_c_name"]
	value, err := fieldType.FieldValueFunc(p)
	if err != nil {
		panic(err)
	}
	fmt.Println(value)

	//for k, v := range proto {
	//	fmt.Printf("k: %v v: %v\n", k, v)
	//	value, err := v.FieldValueFunc(p)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//	} else {
	//		fmt.Println(value)
	//	}
	//}
}

func TestConvertProto(t *testing.T) {
	logger.SetLevel("debug")
	logger.SetLevel("debug")
	name := "test"
	age := 18
	p := &A{
		B: B{
			C: C{
				Name: name,
				Age:  age,
			},
		},
	}
	parser := NewParser(p)
	proto, err := parser.Parse()
	fmt.Printf("fields: %v err: %v\n", proto, err)

	fieldType := proto["b_c_name"]
	value, err := fieldType.FieldValueFunc(p)
	if err != nil {
		t.Fatalf(err.Error())
	}
	fmt.Println(fieldType.Kind)
	fmt.Println(value)
	assert.Equal(t, value, name)
	ageV := proto["b_c_age"]
	if valueFunc, err := ageV.FieldValueFunc(p); err != nil {
		t.Fatalf(err.Error())
	} else {
		assert.Equal(t, valueFunc, age)
	}

	adultV := proto["b_c_adult"]
	if valueFunc, err := adultV.FieldValueFunc(p); err != nil {
		t.Fatalf(err.Error())
	} else {
		fmt.Println(valueFunc)
	}

	nullableV := proto["b_c_nullable"]
	if valueFunc, err := nullableV.FieldValueFunc(p); err != nil {
		t.Fatalf(err.Error())
	} else {
		fmt.Println(valueFunc)
		assert.Equal(t, valueFunc, nil)
	}
	//for k, v := range proto {
	//	fmt.Printf("k: %v v: %v\n", k, v)
	//	value, err := v.FieldValueFunc(p)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//	} else {
	//		fmt.Println(value)
	//	}
	//}
}
