package queue

import (
	"context"
	mqttpb "github.com/blademainer/commons/pkg/rpc/queue/proto"
	"github.com/golang/protobuf/proto"
	"reflect"
	"time"
)

// message consume from queue
//func (s *defaultServer) handleResponse(message *mqttpb.QueueMessage) (e error) {
//	select {
//	case s.responseMessageChan <- message:
//		return nil
//	default:
//		e := fmt.Errorf("responseCh is full, current size: %v", len(s.responseMessageChan))
//		return e
//	}
//}

// clients await response
func (s *defaultServer) awaitResponse(ctx context.Context, messageId string) (message *mqttpb.QueueMessage, e error) {
	doneCh := make(chan *mqttpb.QueueMessage)
	errorCh := make(chan error)
	timeout := s.invokeTimeout
	deadline, ok := ctx.Deadline()
	if ok && deadline.UnixNano()-time.Now().UnixNano() > 0 {
		timeout = time.Duration(deadline.UnixNano()-time.Now().UnixNano()) * time.Nanosecond
	}
	e = s.watchMessageId(messageId, timeout, doneCh, errorCh)
	if e != nil {
		return nil, e
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case e := <-errorCh:
		return nil, e
	case response := <-doneCh:
		return response, e
	}
}

func (s *defaultServer) parseResponse(handlerType interface{}, method string, queueMessage mqttpb.QueueMessage) (response proto.Message, e error) {
	bytes := queueMessage.Message
	found, e := s.getHandleType(handlerType)
	if e != nil {
		return
	}
	grpcMethod := found.MethodMap[method]
	value := reflect.New(grpcMethod.Out.Elem())
	response = value.Interface().(proto.Message)
	e = proto.Unmarshal(bytes, response)
	return
}
