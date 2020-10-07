package sign

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"github.com/blademainer/commons/pkg/random"
)

//func RsaEncrypt(origData []byte, publicKey []byte) ([]byte, error) {
//	block, _ := pem.Decode(publicKey)
//	if block == nil {
//		return nil, errors.New("public key error")
//	}
//	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
//	if err != nil {
//		return nil, err
//	}
//	pub := pubInterface.(*rsa.PublicKey)
//	return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
//}
//
//// 解密
//func RsaDecrypt(ciphertext []byte, privateKey []byte) ([]byte, error) {
//	block, _ := pem.Decode(privateKey)
//	if block == nil {
//		return nil, errors.New("private key error!")
//	}
//	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
//	if err != nil {
//		return nil, err
//	}
//	return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
//}

var randReader = random.NewRand()

func packageData(originalData []byte, packageSize int) (r [][]byte) {
	var src = make([]byte, len(originalData))
	copy(src, originalData)

	r = make([][]byte, 0)
	if len(src) <= packageSize {
		return append(r, src)
	}
	for len(src) > 0 {
		var p = src[:packageSize]
		r = append(r, p)
		src = src[packageSize:]
		if len(src) <= packageSize {
			r = append(r, src)
			break
		}
	}
	return r
}

func RSAEncrypt(plaintext, key []byte) ([]byte, error) {
	var err error
	var block *pem.Block
	block, _ = pem.Decode(key)
	if block == nil {
		return nil, errors.New("public key error")
	}

	var pubInterface interface{}
	pubInterface, err = x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	var pub = pubInterface.(*rsa.PublicKey)

	var data = packageData(plaintext, pub.N.BitLen()/8-11)
	var cipherData []byte = make([]byte, 0, 0)

	for _, d := range data {
		var c, e = rsa.EncryptPKCS1v15(randReader, pub, d)
		if e != nil {
			return nil, e
		}
		cipherData = append(cipherData, c...)
	}

	return cipherData, nil
}

func RSADecrypt(ciphertext, key []byte) ([]byte, error) {
	var err error
	var block *pem.Block
	block, _ = pem.Decode(key)
	if block == nil {
		return nil, errors.New("private key error")
	}

	var pri *rsa.PrivateKey
	pri, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	var data = packageData(ciphertext, pri.PublicKey.N.BitLen()/8)
	var plainData []byte = make([]byte, 0, 0)

	for _, d := range data {
		var p, e = rsa.DecryptPKCS1v15(randReader, pri, d)
		if e != nil {
			return nil, e
		}
		plainData = append(plainData, p...)
	}
	return plainData, nil
}

func RSAEncryptByKey(key *rsa.PublicKey, plaintext []byte) (cipher []byte, err error) {
	plainTextMax := key.Size() - 11
	count := len(plaintext) / plainTextMax
	if len(plaintext)%plainTextMax > 0 {
		count++
	}
	cipher = make([]byte, count*key.Size())
	start := 0

	chunks := packageData(plaintext, plainTextMax)

	for i := 0; i < len(chunks); i++ {
		//size := len(plaintext)
		//if size > plainTextMax {
		//	size = plainTextMax
		//}
		chunk := chunks[i]

		// 对数据进行切片
		cipherTmp, err := rsa.EncryptPKCS1v15(randReader, key, chunk)
		if err != nil {
			return nil, err
		}
		end := start + len(cipherTmp)
		copy(cipher[start:end], cipherTmp)
		start = end
	}
	return
}

func RSADecryptByKey(key *rsa.PrivateKey, cipher []byte) (plaintext []byte, err error) {
	chunks := packageData(cipher, key.Size())
	plaintext = make([]byte, (key.Size()-11)*len(chunks))
	start := 0
	for i := 0; i < len(chunks); i++ {
		chunk := chunks[i]

		// 对数据进行切片
		plaintextTmp, err := rsa.DecryptPKCS1v15(randReader, key, chunk)
		if err != nil {
			return nil, err
		}
		end := start + len(plaintextTmp)
		copy(plaintext[start:end], plaintextTmp)
		start = end
	}
	plaintext = plaintext[:start]
	return
}
