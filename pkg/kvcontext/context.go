package kvcontext

import (
	"context"
	"sync"
)

// This prevents collisions with keys defined in other packages.
type key int

// mapKey is the key for db. Transaction values in Contexts. It is
// unexported; clients use db.NewContext and db.FromContext
// instead of using this key directly.
var mapKey key

// NewKvContext returns a new Context that init map.
func NewKvContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, mapKey, &sync.Map{})
}

// NewContext returns a new Context that carries value t.
func NewContext(ctx context.Context, t *sync.Map) context.Context {
	return context.WithValue(ctx, mapKey, t)
}

// FromContext returns the Transaction value stored in ctx, if any.
func FromContext(ctx context.Context) (*sync.Map, bool) {
	u, ok := ctx.Value(mapKey).(*sync.Map)
	return u, ok
}

// Put kv to context
func Put(ctx context.Context, key interface{}, value interface{}) (bool, error) {
	kv, ok := FromContext(ctx)
	if !ok {
		return false, &NotKvContextError{Message: "not kvContext, please init by NewKvContext(ctx context.Context)"}
	}
	kv.Store(key, value)
	return true, nil
}

// Del delete key
func Del(ctx context.Context, key interface{}) error {
	kv, ok := FromContext(ctx)
	if !ok {
		return &NotKvContextError{Message: "not kvContext, please init by NewKvContext(ctx context.Context)"}
	}
	kv.Delete(key)
	return nil
}

// Get value of key
func Get(ctx context.Context, key interface{}) (interface{}, error) {
	kv, ok := FromContext(ctx)
	if !ok {
		return nil, &NotKvContextError{Message: "not kvContext, please init by NewKvContext(ctx context.Context)"}
	}
	value, ok := kv.Load(key)
	if !ok {
		return nil, nil
	}
	return value, nil
}
