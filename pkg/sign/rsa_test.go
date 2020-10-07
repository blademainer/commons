package sign

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"github.com/blademainer/commons/pkg/util"
	"github.com/stretchr/testify/assert"
	"sync"
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

func TestRsaEncrypt_Encrypt(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err.Error())
	}
	public := key.PublicKey

	for i := 1; i < 4096; i++ {
		source := make([]byte, i)
		read, err := rand.Read(source)
		if err != nil {
			t.Fatal(err.Error())
		} else if read == 0 {
			t.FailNow()
		}

		cipher, err := RSAEncryptByKey(&public, source)
		if err != nil {
			t.Fatal(err.Error())
		}
		decrypt, err := RSADecryptByKey(key, cipher)
		//cipher, err := rsa.EncryptPKCS1v15(rand.Reader, &public, source)
		if err != nil {
			t.Fatal(err.Error())
		}
		assert.Equal(t, source, decrypt)
		//fmt.Printf("source len: %v cipher: %v decrypt: %v\n", read, len(cipher), len(decrypt))
	}
}

func TestRsaEncrypt_ConcurrentEncrypt(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err.Error())
	}
	public := key.PublicKey

	wg := sync.WaitGroup{}
	for i := 1; i < 4096; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			source := make([]byte, i)
			read, err := rand.Read(source)
			if err != nil {
				t.Fatal(err.Error())
			} else if read == 0 {
				t.FailNow()
			}

			cipher, err := RSAEncryptByKey(&public, source)
			if err != nil {
				t.Fatal(err.Error())
			}
			decrypt, err := RSADecryptByKey(key, cipher)
			//cipher, err := rsa.EncryptPKCS1v15(rand.Reader, &public, source)
			if err != nil {
				t.Fatal(err.Error())
			}
			assert.Equal(t, source, decrypt)
		}()
	}
	wg.Wait()
}

func BenchmarkManager_RsaEncrypt(b *testing.B) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		b.Fatal(err.Error())
	}
	public := key.PublicKey

	for i := 0; i < b.N; i++ {
		source := make([]byte, 512)
		read, err := rand.Read(source)
		if err != nil {
			b.Fatal(err.Error())
		} else if read == 0 {
			b.FailNow()
		}

		cipher, err := RSAEncryptByKey(&public, source)
		if err != nil {
			b.Fatal(err.Error())
		}
		decrypt, err := RSADecryptByKey(key, cipher)
		//cipher, err := rsa.EncryptPKCS1v15(rand.Reader, &public, source)
		if err != nil {
			b.Fatal(err.Error())
		}
		assert.Equal(nil, source, decrypt)
		//fmt.Printf("source len: %v cipher: %v\n", read, len(cipher))
	}
}
