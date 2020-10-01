package field

import (
	"fmt"
	"github.com/coreos/etcd/pkg/testutil"
	"strings"
	"testing"
)

// Person test struct
type Person struct {
	Name   string `form:"name"`
	Age    uint8  `form:"age"`
	Gender int `form:"gender"`
}

// TestUnmarshal test
func TestUnmarshal(t *testing.T) {
	parser := &Parser{Tag: "form", Escape: false, GroupDelimiter: '&', PairDelimiter: '='}
	parser.Tag = "form"
	person := &Person{Name: "zhangsan", Age:18}
	params := make(map[string][]string)
	data := parser.Unmarshal(person, params)
	fmt.Println(data)

}

// TestMarshal test
func TestMarshal(t *testing.T) {
	parser := &Parser{Tag: "form", Escape: false, GroupDelimiter: '&', PairDelimiter: '='}
	parser.Tag = "form"

	person := &Person{Name: "张三", Age: 18}

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
// TestMarshal test
func TestMarshalStr(t *testing.T) {
	parser := &Parser{Tag: "form", Quoted: true, Escape: false, GroupDelimiter: '&', PairDelimiter: '='}
	parser.Tag = "form"

	person := &Person{Name: "张三", Age: 18}

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

// TestMarshal test
func TestIgnoreEmptyValue(t *testing.T) {
	parser := &Parser{Tag: "form", Escape: false, GroupDelimiter: '&', PairDelimiter: '=', IgnoreNilValueField: true}
	person := &Person{Age: 18}
	if b, e := parser.Marshal(person); e == nil {
		s := string(b)
		testutil.AssertTrue(t, !strings.Contains(s, "name"))
	} else {
		t.Fail()
	}

	person.Name = ""
	if b, e := parser.Marshal(person); e == nil {
		s := string(b)
		testutil.AssertTrue(t, !strings.Contains(s, "name"), "name mustn't exists!")
	} else {
		t.Fail()
	}

	person.Name = "张三"
	if b, e := parser.Marshal(person); e == nil {
		s := string(b)
		fmt.Println(s)
		testutil.AssertTrue(t, strings.Contains(s, "name"), "name must exists!")
	} else {
		t.Fail()
	}
}

// TestSort test
func TestSort(t *testing.T) {
	parser := HttpFormParser
	person := &Person{Age: 18, Name: "张三", Gender: 1}
	if b, e := parser.Marshal(person); e == nil {
		s := string(b)
		fmt.Println(s)
		testutil.AssertTrue(t, strings.Index(s, "name") > strings.Index(s, "age"), "age must before name!")
	} else {
		t.Fail()
	}
}

// Benchmark benckmark marshal
func Benchmark(b *testing.B) {
	p := HttpEncodedFormParser
	person := &Person{Age: 18, Name: "张三"}
	for i := 0; i < b.N; i++ {
		_, err := p.Marshal(person)
		if err != nil {
			panic(err.Error())
		}
	}
}

// BenchmarkHttpFormParse benchmark HttpFormParser
func BenchmarkHttpFormParse(b *testing.B) {
	p := HttpFormParser
	person := &Person{Age: 18, Name: "张三"}
	for i := 0; i < b.N; i++ {
		_, err := p.Marshal(person)
		if err != nil {
			panic(err.Error())
		}
	}
}
