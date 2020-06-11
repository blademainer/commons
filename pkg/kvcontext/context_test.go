package kvcontext

import (
	"context"
	"gopkg.in/go-playground/assert.v1"
	"testing"
)

func TestNewKvContext(t *testing.T) {
	kvContext := NewKvContext(context.Background())
	ok, err := Put(kvContext, "hello", "world")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal(ok)
	}
	get, err := Get(kvContext, "hello")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal(ok)
	}
	assert.Equal(t, get, "world")

}
