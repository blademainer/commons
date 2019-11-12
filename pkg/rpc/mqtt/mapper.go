package mqtt

import (
	"github.com/blademainer/commons/pkg/logger"
	"google.golang.org/grpc"
	"reflect"
)

type ServiceMapper struct {
	Service interface{}
	// type
	HandlerType interface{}
}

var serviceMap = make(map[interface{}]*ServiceMapper, 0)

//func Put(mapper *ServiceMapper) {
//	serviceMap[mapper.HandlerType] = mapper
//}
func Put(handlerType interface{}, desc *grpc.ServiceDesc) {
	if handlerType == nil {
		return
	}

	tp := reflect.TypeOf(handlerType)
	if tp.Kind() != reflect.Ptr {
		logger.Errorf("tp.Kind(): %v not pointer", tp.Kind())
		return
	}
	name := tp.Elem().Name()
	logger.Infof("type: %v name: %v", tp, name)
	for i := 0; i < tp.NumMethod(); i++ {
		method := tp.Method(i)
		logger.Infof("type: %v method: %v", tp, method)
	}

	serviceMap[handlerType] = &ServiceMapper{HandlerType: handlerType}

}

func Get(handlerType interface{}) *ServiceMapper {
	return serviceMap[handlerType]
}
