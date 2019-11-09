package conf

import (
	"github.com/blademainer/commons/pkg/logger"
	"reflect"
)

func InitConfig(uri string, configInstance interface{}){
	if configInstance != nil && reflect.TypeOf(configInstance).Kind() != reflect.Ptr {
		logger.Errorf("Config instance is not ptr kind!")
		panic("Config instance is not ptr kind!")
	}
}
