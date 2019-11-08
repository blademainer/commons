package sign

import "testing"

func TestFingerprintString(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{`-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAuBgi30BGxlsiwVKnIUgE
dEkXhSPDjktegwl/6NH3zLw6N+dV1kyjGF23YxhBPccnAI10x5BhUr24JeEQx0iQ
bfx+YgkcbWkR87DpctQ3hTW2DPf5UtD5DLAw0FPvKcTDnLynq9s6kpx2ErqxUGgJ
krZEK/El/ViEhvrfyPTwQ2aK4eqjHZQAfk5lew7CoNWH/9FR+HSpZraP+dAKXkOJ
8mAxR/vQmn7/gMiDHLdcELB3SXagoPKPjln5UmcgNbxg/vaLkCx/TLWSdTFrUypa
02+fay2fLaiUGYYHpdHFcQ36HfumdRH79pe5gflmlpeEpac4Kdo9RlNIuRy7Y/pZ
6wIDAQAB
-----END PUBLIC KEY-----`}, "07:4a:d9:ec:83:7d:1e:f6:67:7f:8c:42:dc:7a:79:64"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FingerprintString(tt.args.key); got != tt.want {
				t.Errorf("FingerprintString() = %v, want %v", got, tt.want)
			}
		})
	}
}
