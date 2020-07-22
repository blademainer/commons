package queue

import (
	"context"
	"fmt"
	"github.com/blademainer/commons/pkg/generator"
	"github.com/blademainer/commons/pkg/logger"
	mqttpb "github.com/blademainer/commons/pkg/rpc/queue/proto"
	"github.com/golang/protobuf/proto"
	"reflect"
	"runtime"
	"strings"
	"sync"
)

type Server interface {
	// prefix mqtt listen to
	RegisterService(handlerType interface{}, service interface{}) error
	//RegisterServiceFn(f RegisterFunc)
	Serve() error
	// invokeRequest grpc queue.
	// returns SkipAwaitError when awaitResponse option is off. please use func: queue.IsSkipAwaitError(e) to judgement the error type.
	Invoke(handlerType interface{}, method string, ctx context.Context, message proto.Message) (response proto.Message, messageId *string, e error)
	InvokeWithOpts(handlerType interface{}, method string, ctx context.Context, message proto.Message, opts *InvokeOptions) (response proto.Message, messageId *string, e error)
	Handle(payload []byte) (e error) // handle queue's message by server
	Stop() (e error)
	WatchResponse() (responseMessageChan <-chan *mqttpb.QueueMessage, e error)
}

type AwaitResponseFunc = func(message *mqttpb.QueueMessage, e error)

type defaultServer struct {
	*generator.Generator
	*Options
	sync.RWMutex

	doneCh              chan struct{}
	responseMessageChan chan *mqttpb.QueueMessage // chan for responses
	requestMessageChan  chan *mqttpb.QueueMessage // chan for responses

	queue         Queue
	serviceMap    map[interface{}]serviceAndMethod // [implements of ServerInterface] -> (methodName -> method)
	cmdMap        map[string]*grpcMethod           // cmd -> ServerInterface's methods
	handleTypeMap map[interface{}]*grpcMethods     // ServerInterface's methods

	keeper *awaitKeeper
}

func (s *defaultServer) WatchResponse() (responseMessageChan <-chan *mqttpb.QueueMessage, e error) {
	if s.Options.awaitResponse {
		e = fmt.Errorf("failed to get responseMessageChan, because Options.awaitResponse is on")
		return nil, e
	}
	return s.responseMessageChan, nil
}

func NewServer(client Queue, options *Options) Server {
	server := &defaultServer{}
	server.serviceMap = make(map[interface{}]serviceAndMethod, 0)
	server.cmdMap = make(map[string]*grpcMethod)
	server.handleTypeMap = make(map[interface{}]*grpcMethods)
	server.queue = client
	server.requestMessageChan = make(chan *mqttpb.QueueMessage, options.awaitQueueSize)
	server.responseMessageChan = make(chan *mqttpb.QueueMessage, options.awaitQueueSize)

	if options.awaitResponse {
		keeper := newAwaitKeeper(options, &server.responseMessageChan)
		server.keeper = keeper
		server.startKeeper()
	}
	server.Options = options
	cluster := ""
	g := generator.New(cluster, 1000000)
	server.Generator = g
	return server
}

type serviceAndMethod struct {
	service interface{}
	method  map[string]reflect.Value
}

func (s *defaultServer) Stop() (e error) {
	s.doneCh <- struct{}{}
	close(s.doneCh)
	return nil
}

func (s *defaultServer) Produce(payload []byte) error {
	e := s.queue.Produce(payload)
	return e
}

func (s *defaultServer) Serve() error {
	s.startHandleRequest()
	s.startConsumeQueue()
	return nil
}

func (s *defaultServer) startConsumeQueue() {
	for {
		select {
		case <-s.doneCh:
			return
		default:
			payload, e := s.queue.Consume()
			if e != nil {
				logger.Errorf("failed to consume message: %v", e.Error())
				continue
			}
			e = s.Handle(payload)
			if e != nil {
				logger.Errorf("failed to handle message: %v", e.Error())
				continue
			}
		}
	}
}

func (s *defaultServer) startHandleRequest() {
	for i := 0; i < s.messageHandleConcurrentSize; i++ {
		go func() {
			for {
				select {
				case message := <-s.requestMessageChan:
					responseData, e := s.handleRequest(message)
					if e != nil {
						logger.Errorf("failed to handle message: %v, error: %v", message, e.Error())
					} else if responseData == nil {
						logger.Warnf("failed to handle message: %v because response is nil", message)
					}
					e = s.queue.Produce(responseData)
					if e != nil {
						logger.Errorf("failed to produce message, error: %v", e.Error())
					}
				case <-s.doneCh:
					logger.Warnf("stopping HandleRequest...")
					return
				}
			}
		}()
	}
}

func PrintStack() {
	var buf [4096]byte
	n := runtime.Stack(buf[:], false)
	fmt.Printf("==> %s\n", string(buf[:n]))
}

// RegisterService registers a service and its implementation to the gRPC
// defaultServer. It is called from the IDL generated code. This must be called before
// invoking Serve.
func (s *defaultServer) RegisterService(handlerType interface{}, service interface{}) error {
	ht := reflect.TypeOf(handlerType).Elem()
	st := reflect.TypeOf(service)
	if !st.Implements(ht) {
		PrintStack()
		logger.Fatalf("grpc: Server.RegisterService found the handler of type %v that does not satisfy %v", st, ht)
	}
	//mapper := Get(handlerType)
	//mapper.Service = service
	//reflect.TypeOf(service).MethodByName()
	//s.server.RegisterService(mapper.Desc, service)
	logger.Infof("register service: %v type: %v", st, ht)
	s.Lock()
	defer s.Unlock()
	methods, err := parseService(ht)
	if err != nil {
		return err
	}
	for _, method := range methods.Methods {
		cmd := method.Url()
		s.cmdMap[cmd] = method

		m := reflect.ValueOf(service).MethodByName(method.Name)
		methodNameMap := make(map[string]reflect.Value)
		methodNameMap[method.Name] = m
		s.serviceMap[ht] = serviceAndMethod{
			service: service,
			method:  methodNameMap,
		}
	}
	return nil
}

func (s *defaultServer) Handle(payload []byte) (e error) {
	if payload == nil {
		e = fmt.Errorf("payload is nil")
		return
	}
	// unmarshal
	message := &mqttpb.QueueMessage{}
	e = proto.Unmarshal(payload, message)
	if e != nil {
		return
	}

	if logger.IsDebugEnabled() {
		logger.Debugf("receive type: %v message: %v", message.Type, message)
	}

	switch message.Type {
	case mqttpb.MessageType_REQUEST:
		s.requestMessageChan <- message
		return nil
	case mqttpb.MessageType_RESPONSE:
		s.responseMessageChan <- message
		return nil
	default:
		logger.Errorf("unknown message type: %v", message.Type)
		return
	}
}

func (s *defaultServer) Invoke(handlerType interface{}, method string, ctx context.Context, message proto.Message) (response proto.Message, messageId *string, e error) {
	options := NewInvokeOptions()
	options.produceFunc = s.Produce
	return s.InvokeWithOpts(handlerType, method, ctx, message, options)
}

func (s *defaultServer) InvokeWithOpts(handlerType interface{}, method string, ctx context.Context, message proto.Message, opts *InvokeOptions) (response proto.Message, messageId *string, e error) {
	if handlerType == nil {
		e = fmt.Errorf("handlerType is nil")
		return
	}
	messageId, e = s.invokeRequest(handlerType, method, ctx, message, opts)
	if e != nil {
		return
	} else if messageId == nil {
		e = fmt.Errorf("handlerType: %v method: %v request: %v no messageId returned", handlerType, method, message)
		return
	}

	if !s.Options.awaitResponse {
		e = &SkipAwaitError{"skip waiting response, because options.awaitResponse is false. please use func: queue.IsSkipAwaitError(e) to judgement the error type."}
		return
	}

	// await response
	queueMessage, e := s.awaitResponse(ctx, *messageId)
	if e != nil {
		return
	} else if queueMessage == nil {
		e = fmt.Errorf("no response returned")
		return
	}
	response, e = s.parseResponse(handlerType, method, *queueMessage)
	return
}

type grpcMethods struct {
	Type      reflect.Type
	Name      string
	Methods   []*grpcMethod
	MethodMap map[string]*grpcMethod
}

type grpcMethod struct {
	TypeOfInterface reflect.Type
	Method          reflect.Method
	Name            string
	In              reflect.Type
	Out             reflect.Type
}

func (g *grpcMethod) String() string {
	return fmt.Sprintf("url: %v in: %v out: %v", g.Url(), g.In, g.Out)
}

func (g *grpcMethod) Url() string {
	split := strings.Split(g.TypeOfInterface.PkgPath(), "/")
	return fmt.Sprintf("/%v.%v/%v", split[len(split)-1], g.TypeOfInterface.Name(), g.Name)
}

func BuildUrl(interfaceType reflect.Type, method reflect.Method) string {
	split := strings.Split(interfaceType.PkgPath(), "/")
	return fmt.Sprintf("/%v.%v/%v", split[len(split)-1], interfaceType.Name(), method.Name)
}

func (s *defaultServer) getHandleType(handlerType interface{}) (methods *grpcMethods, e error) {
	s.RLock()
	defer s.RUnlock()
	methods, exists := s.handleTypeMap[handlerType]
	ht := reflect.TypeOf(handlerType).Elem()
	if !exists {
		//       code block start								       code block end
		//             ⬇													  ⬇
		// ReadLock -> {  ReadUnlock -> Lock -> write() -> UnLock -> ReadLock } -> ReadUnlock

		// release ReadLock for next step's write lock
		s.RUnlock()
		// get ReadLock before the defer's last ReadUnlock
		defer s.RLock()
		// begin write
		s.Lock()
		// release write lock
		defer s.Unlock()
		if methods, exists = s.handleTypeMap[handlerType]; exists {
			return
		}
		grpcMethods, e := parseService(ht)
		logger.Infof("parse service: %v to grpcMethods: %v", ht, grpcMethods)
		if e != nil {
			return methods, e
		}
		s.handleTypeMap[handlerType] = grpcMethods
		methods = grpcMethods
	}
	return
}

func parseService(ht reflect.Type) (*grpcMethods, error) {
	name := ht.Name()
	path := ht.PkgPath()
	logger.Infof("type: %v name: %v path: %v", ht.String(), name, path)

	methodMap := make(map[string]*grpcMethod)
	methods := make([]*grpcMethod, 0)
	for i := 0; i < ht.NumMethod(); i++ {
		method := ht.Method(i)
		logger.Infof("type: %v method: %v", ht, method)
		mt := method.Type
		ni := mt.NumIn()
		if ni != 2 {
			return nil, fmt.Errorf("method: %v is not 2 size of args, actual number: %v", method, ni)
		}
		arg0 := mt.In(0)
		if arg0 != reflect.TypeOf((*context.Context)(nil)).Elem() {
			return nil, fmt.Errorf("method: %v arg0 is not *context.Context type, actual type: %v", method, arg0)
		}
		arg1 := mt.In(1)
		logger.Infof("method: %v arg1 type type: %v", method, arg1)
		if !isArgsImplementsProtoMessage(arg1) {
			return nil, fmt.Errorf("arg1: %v is not implements proto.Message", arg1)
		}

		no := mt.NumOut()
		if no != 2 {
			return nil, fmt.Errorf("method: %v is not 2 size of out, actual number: %v", method, no)
		}
		out0 := mt.Out(0)
		if !isArgsImplementsProtoMessage(out0) {
			return nil, fmt.Errorf("return result: %v is not implements proto.Message", out0)
		}
		logger.Infof("method: %v out0 type: %v", method, out0)
		out1 := mt.Out(1)
		logger.Infof("method: %v out1 type: %v", method, out1)
		if out1 != reflect.TypeOf((*error)(nil)).Elem() {
			return nil, fmt.Errorf("method: %v arg0 is not context type, actual type: %v", method, out1)
		}

		grpcMethod := &grpcMethod{}
		grpcMethod.TypeOfInterface = ht
		grpcMethod.Name = method.Name
		grpcMethod.Method = method
		grpcMethod.Out = out0
		grpcMethod.In = arg1
		methods = append(methods, grpcMethod)

		methodMap[method.Name] = grpcMethod
	}

	grpcMethods := &grpcMethods{
		Methods:   methods,
		Type:      ht,
		Name:      ht.Name(),
		MethodMap: methodMap,
	}

	return grpcMethods, nil
}

func isArgsImplementsProtoMessage(arg reflect.Type) bool {
	messageType := reflect.TypeOf((*proto.Message)(nil)).Elem()
	return arg.Implements(messageType)
}
