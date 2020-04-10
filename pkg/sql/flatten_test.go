package sql

import (
	"encoding/json"
	"fmt"
	"github.com/blademainer/commons/pkg/logger"
	"gotest.tools/assert"
	"testing"
)

type A struct {
	B B
	C *C
}

type B struct {
	C C
}

type C struct {
	Name       string
	Age        int
	Adult      bool
	Nullable   *string
	Labels     map[string]string
	LabelsPtr  *map[string]string
	LabelArray []string
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
	parser, err := NewParser(p)
	proto := parser.GetFields()
	fmt.Printf("fields: %v err: %v\n", proto, err)

	// name field of A.B.C.Name
	fieldEntry := proto["b_c_name"]
	value, err := fieldEntry.fieldValueFunc(p)
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

func TestConvertJson(t *testing.T) {
	logger.SetLevel("debug")
	name := "test"
	age := 18
	labels := make(map[string]string)
	labels["region"] = "cn"
	labels["zone"] = "sz"

	labelArray := []string{"a", "b"}

	p := A{
		B: B{
			C: C{
				Name:       name,
				Age:        age,
				Labels:     labels,
				LabelArray: labelArray,
			},
		},
	}
	parser, err := NewParser(p)
	fields := parser.GetFields()
	for s, entry := range fields {
		fmt.Printf("k: %v v: %v\n", s, entry)
	}

	// name field of A.B.C.Name
	fieldEntry := fields["b_c_name"]
	value, err := fieldEntry.fieldValueFunc(p)
	if err != nil {
		t.Fatalf(err.Error())
	}
	fmt.Println("type: ", fieldEntry.Type)
	fmt.Println("fieldEntry: ", fieldEntry)
	fmt.Println(value)
	assert.Equal(t, value, name)
	ageV := fields["b_c_age"]
	if valueFunc, err := ageV.fieldValueFunc(p); err != nil {
		t.Fatalf(err.Error())
	} else {
		assert.Equal(t, valueFunc, age)
	}

	adultV := fields["b_c_adult"]
	if valueFunc, err := adultV.fieldValueFunc(p); err != nil {
		t.Fatalf(err.Error())
	} else {
		fmt.Println(valueFunc)
	}

	nullableV := fields["b_c_nullable"]
	if valueFunc, err := nullableV.fieldValueFunc(p); err != nil {
		t.Fatalf(err.Error())
	} else {
		fmt.Println(valueFunc)
		assert.Equal(t, valueFunc, nil)
	}

	e, found := fields["b_c"]
	assert.Equal(t, found, true)
	assert.Equal(t, e.FieldType, Node)
	//for k, v := range fields {
	//	fmt.Printf("k: %v v: %v\n", k, v)
	//	value, err := v.FieldValueFunc(p)
	//	if err != nil {
	//		fmt.Println(err.Error())
	//	} else {
	//		fmt.Println(value)
	//	}
	//}
}

func Benchmark(b *testing.B) {
	for i := 0; i < b.N; i++ {
		name := "test"
		age := 18
		p := A{
			B: B{
				C: C{
					Name: name,
					Age:  age,
				},
			},
		}
		parser, err := NewParser(p)
		if err != nil {
			panic(err.Error())
		}
		proto := parser.GetFields()

		// name field of A.B.C.Name
		fieldEntry := proto["b_c_name"]
		fieldEntry.fieldValueFunc(p)
	}

}

func BenchmarkFieldValueFunc(b *testing.B) {
	name := "test"
	age := 18
	p := A{
		B: B{
			C: C{
				Name: name,
				Age:  age,
			},
		},
	}
	parser, err := NewParser(p)
	if err != nil {
		panic(err.Error())
	}
	proto := parser.GetFields()
	for i := 0; i < b.N; i++ {
		// name field of A.B.C.Name
		fieldEntry := proto["b_c_name"]
		fieldEntry.fieldValueFunc(p)
	}
}

func BenchmarkReflectParser_GetValueMap(b *testing.B) {
	name := "test"
	age := 18
	p := A{
		B: B{
			C: C{
				Name: name,
				Age:  age,
			},
		},
	}
	parser, err := NewParser(p)
	if err != nil {
		panic(err.Error())
	}
	for i := 0; i < b.N; i++ {
		// name field of A.B.C.Name
		parser.GetValueMap(p)
	}
}

func Test_reflectParser_GetValueMap(t *testing.T) {
	logger.SetLevel("debug")
	name := "test"
	age := 18
	labels := make(map[string]string)
	labels["region"] = "cn"
	labels["zone"] = "sz"
	labelArray := []string{"a", "b"}

	p := A{
		B: B{
			C: C{
				Name:       name,
				Age:        age,
				Labels:     labels,
				LabelArray: labelArray,
				LabelsPtr:  &labels,
			},
		},
	}
	parser, err := NewParser(p)
	valueMap, err := parser.GetValueMap(&p)
	if err != nil {
		t.Fatal(err.Error())
	}
	for k, v := range valueMap {
		fmt.Printf("k: %v v: %v\n", k, v)
	}
	assert.Equal(t, valueMap["b_c_name"], name)
	assert.Equal(t, valueMap["b_c_age"], age)
}

func Test_reflectParser_ResolveFieldsFromMap(t *testing.T) {
	logger.SetLevel("debug")
	name := "test"
	age := 18
	cName := "cname"
	cAge := 20
	labels := make(map[string]string)
	labels["region"] = "cn"
	labels["zone"] = "sz"

	labelArray := []string{"a", "b"}

	p := A{
		B: B{
			C: C{
				Name:       name,
				Age:        age,
				Labels:     labels,
				LabelArray: labelArray,
				LabelsPtr:  &labels,
			},
		},
		C: &C{
			Name: cName,
			Age:  cAge,
		},
	}
	parser, err := NewParser(p)
	valueMap, err := parser.GetValueMap(&p)
	if err != nil {
		t.Fatal(err.Error())
	}

	v := &A{}
	err = parser.ResolveFieldsFromMap(valueMap, v)
	marshal, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Printf("conver map: %v to instance: %v\n", valueMap, string(marshal))
	assert.Equal(t, v.C.Age, cAge)
	assert.Equal(t, v.C.Name, cName)
	assert.Equal(t, v.B.C.Name, name)
	assert.Equal(t, v.B.C.Age, age)
	assert.Equal(t, v.B.C.Labels["region"], "cn")
	assert.Equal(t, v.B.C.Labels["zone"], "sz")
	assert.Equal(t, v.B.C.LabelArray[0], "a")
	assert.Equal(t, v.B.C.LabelArray[1], "b")

	c2 := C{}
	err = parser.ResolveFieldsFromMap(valueMap, c2)
	assert.Equal(t, err != nil, true)
	fmt.Println("error: ", err.Error())
}

func BenchmarkReflectParser_ResolveFieldsFromMap(b *testing.B) {
	name := "test"
	age := 18
	cName := "cname"
	cAge := 20
	p := A{
		B: B{
			C: C{
				Name: name,
				Age:  age,
			},
		},
		C: &C{
			Name: cName,
			Age:  cAge,
		},
	}
	parser, err := NewParser(p)
	valueMap, err := parser.GetValueMap(&p)
	if err != nil {
		panic(err)
	}

	v := &A{}
	for i := 0; i < b.N; i++ {
		err = parser.ResolveFieldsFromMap(valueMap, v)
	}
}
