package conf

import (
	"fmt"
	"reflect"
	"testing"
)

type a struct {
	m1 map[string]b
	m2 map[string]*b
	m3 *map[string]b
	bb b
}

type b struct {
	name string
	age uint32
}

func TestType(t *testing.T) {
	ai := &a{}
	ai.m1 = make(map[string]b)
	ai.m2 = make(map[string]*b)
	bs := make(map[string]b)
	ai.m3 = &bs
	ai.bb = b{}

	of := reflect.ValueOf(ai)
	fmt.Println("Type: ", of.Type())
	fmt.Println("Kind: ", of.Type().Kind())
	fmt.Println("Value: ", of)
}

