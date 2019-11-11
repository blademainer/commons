package sign

import (
	"bytes"
	"crypto"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"github.com/blademainer/commons/pkg/logger"
	"github.com/blademainer/commons/pkg/util"
)

func init() {
	initCheckSignMap()
}

type CheckSignValidator struct {
	paramsCompacter ParamsCompacter
}

type CheckSignInterface interface {
	checkSign(source []byte, signMsg string, key string) error
	sign(source []byte, key string) (string, error)
	signType() string
}

var checkSignMap = make(map[string]CheckSignInterface)

func initCheckSignMap() {
	checkSignMap[SIGN_TYPE_MD5] = &Md5{}
	checkSignMap[SIGN_TYPE_SHA256_WITH_RSA] = &Sha256WithRSA{}
}

type Md5 struct {
}

func (m *Md5) sign(source []byte, key string) (string, error) {
	buffer := bytes.NewBuffer(source)
	buffer.Write([]byte(key))
	b := buffer.Bytes()
	sum := md5.Sum(b)
	s := hex.EncodeToString(sum[:])
	return s, nil
}

func (m *Md5) checkSign(source []byte, signMsg string, key string) error {
	generated, e := m.sign(source, key)
	if e != nil {
		logger.Errorf("Failed to generate sign! error: %v", e.Error())
		return e
	}
	if !util.EqualsIgnoreCase(generated, signMsg) {
		e := errors.New("check sign error")
		logger.Warnf("Failed to check sign! ours: %v actual: %v", generated, signMsg)
		return e
	}

	return nil
}

func (*Md5) signType() string {
	return SIGN_TYPE_MD5
}

type Sha256WithRSA struct {
}

func (s *Sha256WithRSA) sign(source []byte, key string) (sign string, err error) {
	signBytes, err := SignPKCS1v15WithStringKey(source, key, crypto.SHA256)
	if err != nil {
		logger.Errorf("Failed to sign! error: %v key: %v", err.Error(), key)
		return
	}
	sign = base64.StdEncoding.EncodeToString(signBytes)
	logger.Debugf("Encode source: %v to sign: %v", string(source), sign)
	return
}

func (*Sha256WithRSA) checkSign(source []byte, signMsg string, key string) (err error) {
	sign, err := base64.StdEncoding.DecodeString(signMsg)
	if err != nil {
		logger.Errorf("Failed to check sign! decode sign: %v with error: %v", signMsg, err.Error())
		return
	}
	err = VerifyPKCS1v15WithStringKey(source, sign, key, crypto.SHA256)
	if err != nil {
		logger.Errorf("Failed to check sign! check source: %v sign: %v with error: %v", string(source), signMsg, err.Error())
		return
	}
	return err
}

func (*Sha256WithRSA) signType() string {
	return SIGN_TYPE_SHA256_WITH_RSA
}
