package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReplaceGatewayOrderId(t *testing.T) {
	gatewayOrderId := RandString(64)
	id := ReplaceGatewayOrderId("http://127.0.0.1:8888/notify/{gateway_order_id}", gatewayOrderId)
	assert.Equal(t, "http://127.0.0.1:8888/notify/"+gatewayOrderId, id)
}

func BenchmarkReplaceGatewayOrderId(b *testing.B) {
	for i := 0; i < b.N; i++ {
		gatewayOrderId := RandString(64)
		ReplaceGatewayOrderId("http://127.0.0.1:8888/notify/{gateway_order_id}", gatewayOrderId)
	}
}

func BenchmarkReplacePlaceholder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		gatewayOrderId := RandString(64)
		ReplacePlaceholder("http://127.0.0.1:8888/notify/{gateway_order_id}", "gateway_order_id", gatewayOrderId)
	}
}

func TestReplacePlaceholders(t *testing.T) {
	type args struct {
		urlPattern string
		kv         map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "t1",
			args: args{
				urlPattern: "notify/{a1}/{a2}/{a3}",
				kv: map[string]string{
					"a1": "foo",
					"a2": "bar",
					"a3": "foobar",
				},
			},
			want:    "notify/foo/bar/foobar",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := ReplacePlaceholders(tt.args.urlPattern, tt.args.kv)
				if (err != nil) != tt.wantErr {
					t.Errorf("ReplacePlaceholders() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if got != tt.want {
					t.Errorf("ReplacePlaceholders() got = %v, want %v", got, tt.want)
				}
			},
		)
	}
}
