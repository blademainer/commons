package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func ExampleParseDoc() {
	type Another struct {
		Name string `json:"name" doc:"name=名字,desc=测试1,required=true"`
	}

	type Person struct {
		Name    string  `json:"name" doc:"name=名字,desc=测试1,required=true"`
		Age     int     `json:"age" doc:"name=年龄,desc=测试2,required=true"`
		Married bool    `json:"married" doc:"name=是否已婚,desc=测试3,required=true"`
		Another Another `json:"another" doc:"name=同伴,desc=测试4,required=false"`
	}

	p := &Person{
		Name:    "zhangsan",
		Age:     18,
		Married: false,
	}
	got, err := ParseDoc(p)
	if err != nil {
		panic(err)
	}
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "    ")
	if err := enc.Encode(got); err != nil {
		panic(err)
	}

	fmt.Println(string(buf.Bytes()))
	//{
	//    "fields": [
	//        {
	//            "name": "名字",
	//            "desc": "测试1",
	//            "required": "true",
	//            "remark": "",
	//            "type": "string",
	//            "doc": null
	//        },
	//        {
	//            "name": "年龄",
	//            "desc": "测试2",
	//            "required": "true",
	//            "remark": "",
	//            "type": "int",
	//            "doc": null
	//        },
	//        {
	//            "name": "是否已婚",
	//            "desc": "测试3",
	//            "required": "true",
	//            "remark": "",
	//            "type": "bool",
	//            "doc": null
	//        },
	//        {
	//            "name": "同伴",
	//            "desc": "测试4",
	//            "required": "false",
	//            "remark": "",
	//            "type": "struct",
	//            "doc": {
	//                "fields": [
	//                    {
	//                        "name": "名字",
	//                        "desc": "测试1",
	//                        "required": "true",
	//                        "remark": "",
	//                        "type": "string",
	//                        "doc": null
	//                    }
	//                ],
	//                "demo_json": "{\"name\":\"\"}"
	//            }
	//        }
	//    ],
	//    "demo_json": "{\"name\":\"zhangsan\",\"age\":18,\"married\":false,\"another\":{\"name\":\"\"}}"
	//}
	//
}

func TestParseDoc(t *testing.T) {
	type Another struct {
		Name string `json:"name" doc:"name=名字,desc=测试1,required=true"`
	}

	type Person struct {
		Name    string  `json:"name" doc:"name=名字,desc=测试1,required=true"`
		Age     int     `json:"age" doc:"name=年龄,desc=测试2,required=true"`
		Married bool    `json:"married" doc:"name=是否已婚,desc=测试3,required=true"`
		Another Another `json:"another" doc:"name=同伴,desc=测试4,required=false"`
	}

	p := &Person{
		Name:    "zhangsan",
		Age:     18,
		Married: false,
	}
	got, err := ParseDoc(p)
	if err != nil {
		panic(err)
	}
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "    ")
	if err := enc.Encode(got); err != nil {
		panic(err)
	}
	fmt.Println(string(buf.Bytes()))
}
