package util

import (
	"fmt"
	"testing"
)

func TestGetMacAddrs(t *testing.T) {
	fmt.Printf("mac addrs: %q\n", GetMacAddrs())
	fmt.Printf("ips: %q\n", GetIPs())
}
