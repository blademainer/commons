package sign

import (
	"fmt"
	"github.com/blademainer/commons/pkg/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSign(t *testing.T) {
	generator, _ := NewRsaGenerator(2048)
	pemPrivatePKCS1Key, _ := generator.GeneratePemPrivatePKCS1Key()
	fmt.Println("Private PKCS1 key pem: ", pemPrivatePKCS1Key)
	pemPublicPKIXKey, _ := generator.GeneratePemPublicPKIXKey()
	fmt.Println("Public PKIX key pem: ", pemPublicPKIXKey)

	plainText := util.RandString(10240)
	fmt.Println("plainText: ", plainText)
	if bytes, e := RSAEncrypt([]byte(plainText), []byte(pemPublicPKIXKey)); e != nil {
		fmt.Println("Encrypt Error: ", e.Error())
		assert.Nil(t, e)
	} else {
		if decrypt, err := RSADecrypt(bytes, []byte(pemPrivatePKCS1Key)); err != nil {
			assert.Nil(t, err)
		} else {
			result := string(decrypt)
			fmt.Println("Decrypt: ", result)
			assert.Equal(t, plainText, result)
		}
	}
}
