package sign

import (
	"crypto/md5"
	"encoding/pem"
	"fmt"
	"log"
	"strings"
)

func FingerprintString(key string) string {
	return Fingerprint([]byte(key))
}

func Fingerprint(key []byte) string {
	parts := strings.Fields(string(key))
	if len(parts) < 2 {
		log.Fatal("bad key")
	}

	p, _ := pem.Decode(key)

	//k, err := base64.StdEncoding.DecodeString(parts[1])
	//if err != nil {
	//	log.Fatal(err)
	//}

	fp := md5.Sum(p.Bytes)
	rs := strings.Builder{}
	for i, b := range fp {
		rs.WriteString(fmt.Sprintf("%02x", b))
		if i < len(fp)-1 {
			rs.WriteString(":")
		}
	}
	return rs.String()
}
