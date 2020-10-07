package field

import (
	"fmt"
	"gopkg.in/go-playground/assert.v1"
	"strings"
	"testing"
)

type Person struct {
	Name string `form:"name"`
	Age  uint8  `form:"age"`
}

func TestUnmarshal(t *testing.T) {
	parser := &Parser{Tag: "form", Escape: false, GroupDelimiter: '&', PairDelimiter: '='}
	parser.Tag = "form"
	person := &Person{"zhangsan", 18}
	params := make(map[string][]string)
	data := parser.Unmarshal(person, params)
	fmt.Println(data)

}

func TestMarshal(t *testing.T) {
	parser := &Parser{Tag: "form", Escape: false, GroupDelimiter: '&', PairDelimiter: '='}
	parser.Tag = "form"

	person := &Person{"张三", 18}

	if b, e := parser.Marshal(person); e == nil {
		fmt.Println(string(b))
	} else {
		t.Fail()
	}

	m := map[string]string{"a": "b", "你好": "呵呵"}

	if b, e := parser.Marshal(m); e == nil {
		fmt.Println(string(b))
	} else {
		t.Fail()
	}
}

func TestIgnoreEmptyValue(t *testing.T) {
	parser := &Parser{Tag: "form", Escape: false, GroupDelimiter: '&', PairDelimiter: '=', IgnoreNilValueField: true}
	person := &Person{Age: 18}
	if b, e := parser.Marshal(person); e == nil {
		s := string(b)
		assert.Equal(t, strings.Index(s, "name") < 0, true)
	} else {
		t.Fail()
	}

	person.Name = ""
	if b, e := parser.Marshal(person); e == nil {
		s := string(b)
		assert.Equal(t, strings.Index(s, "name") < 0, true)
	} else {
		t.Fail()
	}

	person.Name = "张三"
	if b, e := parser.Marshal(person); e == nil {
		s := string(b)
		fmt.Println(s)
		assert.Equal(t, strings.Index(s, "name") >= 0, true)
	} else {
		t.Fail()
	}
}

func TestSort(t *testing.T) {
	parser := Parser{IgnoreNilValueField: true, Sort: true, Tag: "form", Escape: true, GroupDelimiter: '&', PairDelimiter: '='}
	person := &Person{Age: 18, Name: "张三"}
	if b, e := parser.Marshal(person); e == nil {
		s := string(b)
		fmt.Println(s)
		assert.Equal(t, strings.Index(s, "name") > strings.Index(s, "age"), true)
	} else {
		t.Fail()
	}
}

func Benchmark(b *testing.B) {
	p := HTTP_ENCODED_FORM_PARSER
	person := &Person{Age: 18, Name: "张三"}
	for i := 0; i < b.N; i++ {
		p.Marshal(person)
	}
}
