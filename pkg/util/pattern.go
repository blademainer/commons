package util

import (
	"fmt"
	"github.com/blademainer/commons/pkg/logger"
	"regexp"
	"strings"
)

var regexMap = make(map[string]*regexp.Regexp)

func ReplaceGatewayOrderId(urlPattern string, gatewayOrderId string) string {
	if url, e := ReplacePlaceholder(urlPattern, "gateway_order_id", gatewayOrderId); e != nil {
		return strings.Replace(urlPattern, "{gateway_order_id}", gatewayOrderId, -1)
	} else {
		return url
	}
}

func ReplacePlaceholder(urlPattern string, placeHolderName string, parameter string) (string, error) {
	if compile, e := GetPlaceholderRegex(placeHolderName); e != nil {
		logger.Errorf("Regex error: %v", e.Error())
		return "", e
	} else {
		result := compile.ReplaceAll([]byte(urlPattern), []byte(parameter))
		return string(result), nil
	}
}

func GetPlaceholderRegex(placeHolderName string) (*regexp.Regexp, error) {
	if regex, found := regexMap[placeHolderName]; found {
		return regex, nil
	}
	pattern := fmt.Sprintf("\\{\\s*%s\\s*\\}", placeHolderName)
	if compile, e := regexp.Compile(pattern); e != nil {
		logger.Errorf("Regex error: %v", e.Error())
		return nil, e
	} else {
		regexMap[placeHolderName] = compile
		return compile, e
	}
}
