package queue

import (
	"context"
	"fmt"
	mqttpb "github.com/blademainer/commons/pkg/rpc/queue/proto"
	"github.com/golang/protobuf/proto"
	"github.com/blademainer/commons/pkg/logger"
	"reflect"
)

func (s *defaultServer) handleRequest(message *mqttpb.QueueMessage) (responseData []byte, e error) {
	// find method
	cmd := message.Command
	bytes := message.Message

	s.RLock()
	defer s.RUnlock()

	method, exists := s.cmdMap[cmd]
	if !exists {
		e = fmt.Errorf("could't found method of cmd: %v please use RegisterService to register your server", cmd)
		return
	}
	ht := method.TypeOfInterface
	service, exists := s.serviceMap[ht]
	//service.method.Func.Call([]reflect.Value{context.Background(), message})
	if !exists {
		e = fmt.Errorf("could't found service of cmd: %v please use RegisterService to register your type: %v", cmd, ht)
		return
	}
	value := reflect.New(method.In.Elem())
	grpcMessage := value.Interface().(proto.Message)
	e = proto.Unmarshal(bytes, grpcMessage)
	if e != nil {
		logger.Errorf("failed to unmarshal message to proto: %v", value)
		return
	}

	serviceMethod, exists := service.method[method.Name]
	if !exists {
		e = fmt.Errorf("could't found service method of cmd: %v please use RegisterService to register your server", cmd)
		return
	}

	// invokeRequest
	ctx, cancelFunc := context.WithTimeout(context.Background(), s.invokeTimeout)
	defer cancelFunc()
	values := invoke(serviceMethod, ctx, grpcMessage)
	if logger.IsDebugEnabled() {
		logger.Debugf("succeed to invokeRequest method: %v and returns: %v", serviceMethod, values)
	}
	if len(values) != 2 {
		e = fmt.Errorf("return values length of method: %v is not 2", serviceMethod)
		return
	}

	// response
	response := &mqttpb.QueueMessage{}
	response.Type = mqttpb.MessageType_RESPONSE
	response.MessageId = message.MessageId

	if values[1].Interface() != nil {
		err, cast := values[1].Interface().(error)
		if !cast {
			e = fmt.Errorf("failed to cast return values[1]: %v to error", values[1])
			logger.Error(e.Error())
			return
		} else if err != nil {
			if logger.IsDebugEnabled() {
				logger.Debugf("invoke method: %v and returns error: %v", cmd, err.Error())
			}
			response.Command = cmd
			response.Error = err.Error()
			response.Success = false
			responseData, e = proto.Marshal(response)
			return
		}
	}

	// marshal data
	responseMessage := values[0].Interface().(proto.Message)
	responseMessageData, e := proto.Marshal(responseMessage)
	if e != nil {
		return
	}
	response.Message = responseMessageData

	responseData, e = proto.Marshal(response)
	return
}

func invoke(method reflect.Value, args ...interface{}) []reflect.Value {
	inputs := make([]reflect.Value, len(args))
	for i, _ := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}
	out := method.Call(inputs)
	return out
}

func (s *defaultServer) invokeRequest(handlerType interface{}, method string, ctx context.Context, message proto.Message, options *InvokeOptions) (messageId *string, e error) {
	ht := reflect.TypeOf(handlerType).Elem()
	found, e := s.getHandleType(handlerType)
	if e != nil {
		return
	}
	m, methodExists := found.MethodMap[method]
	if !methodExists {
		e := fmt.Errorf("could'nt found method: %v on type: %v. exists methods: %v", method, ht, found.MethodMap)
		return nil, e
	}
	bytes, e := proto.Marshal(message)
	if e != nil {
		return nil, e
	}
	if logger.IsDebugEnabled() {
		logger.Debugf("marshaled message: %v to: %v", message, bytes)
	}
	cmd := BuildUrl(ht, m.Method)
	mqttMessage := &mqttpb.QueueMessage{}
	mqttMessage.Command = cmd
	mqttMessage.Message = bytes
	mqttMessage.MessageId = s.GenerateId()
	mqttMessage.Type = mqttpb.MessageType_REQUEST

	messageId = &mqttMessage.MessageId
	raw, e := proto.Marshal(mqttMessage)
	if e != nil {
		return nil, e
	}

	if logger.IsDebugEnabled() {
		logger.Debugf("marshaled mqttMessage: %v to: %v", mqttMessage, raw)
	}
	e = options.produceFunc(raw)
	return
}
