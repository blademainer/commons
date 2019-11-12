package queue

import (
	"context"
	"fmt"
	"github.com/blademainer/commons/pkg/logger"
	mqttpb "github.com/blademainer/commons/pkg/rpc/queue/proto"
	"github.com/golang/protobuf/proto"
	"reflect"
	"strings"
	"time"
)

type Server interface {
	// prefix mqtt listen to
	SetTopic(topic string)
	RegisterService(handlerType interface{}, service interface{}) error
	//RegisterServiceFn(f RegisterFunc)
	Serve() error
	Invoke(handlerType interface{}, method string, ctx context.Context, message proto.Message) error // invoke grpc client
	Handle(payload []byte) (e error)                                                                 // handle queue's message
}

type defaultServer struct {
	client        Queue
	topic         string
	serviceMap    map[interface{}]serviceAndMethod // [implements of ServerInterface] -> (methodName -> method)
	cmdMap        map[string]*grpcMethod           // cmd -> ServerInterface's methods
	handleTypeMap map[interface{}]*grpcMethods     // ServerInterface's methods
	invokeTimeout time.Duration
}

type serviceAndMethod struct {
	service interface{}
	method  map[string]reflect.Value
}

func (s *defaultServer) Produce(payload []byte) error {
	e := s.client.Produce(payload)
	return e
}

func (s *defaultServer) Serve() error {
	//TODO start server

	return nil
}

func (s *defaultServer) SetTopic(topic string) {
	s.topic = topic
}

func NewServer(client Queue, prefix string, invokeTimeout time.Duration) Server {
	server := &defaultServer{}
	server.serviceMap = make(map[interface{}]serviceAndMethod, 0)
	server.cmdMap = make(map[string]*grpcMethod)
	server.handleTypeMap = make(map[interface{}]*grpcMethods)
	server.client = client
	server.SetTopic(prefix)
	server.invokeTimeout = invokeTimeout
	return server
}

// RegisterService registers a service and its implementation to the gRPC
// defaultServer. It is called from the IDL generated code. This must be called before
// invoking Serve.
func (s *defaultServer) RegisterService(handlerType interface{}, service interface{}) error {
	ht := reflect.TypeOf(handlerType).Elem()
	st := reflect.TypeOf(service)
	if !st.Implements(ht) {
		logger.Fatalf("grpc: Server.RegisterService found the handler of type %v that does not satisfy %v", st, ht)
	}
	//mapper := Get(handlerType)
	//mapper.Service = service
	//reflect.TypeOf(service).MethodByName()
	//s.server.RegisterService(mapper.Desc, service)
	logger.Infof("register service: %v type: %v", st, ht)
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

// parse cmd -> server.method and cmd -> service.method
func (s *defaultServer) parseCmdAndServiceMethod() {

}

func (s *defaultServer) Handle(payload []byte) (e error) {
	if payload == nil {
		e = fmt.Errorf("payload is nil")
		return
	}
	message := &mqttpb.QueueMessage{}
	e = proto.Unmarshal(payload, message)
	if e != nil {
		return e
	}
	cmd := message.Command
	bytes := message.Message
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
		return e
	}

	serviceMethod, exists := service.method[method.Name]
	if !exists {
		e = fmt.Errorf("could't found service method of cmd: %v please use RegisterService to register your server", cmd)
		return
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), s.invokeTimeout)
	defer cancelFunc()
	values := invoke(serviceMethod, ctx, grpcMessage)
	if logger.IsDebugEnabled() {
		logger.Debugf("succeed to invoke method: %v and returns: %v", serviceMethod, values)
	}
	return nil
}

func invoke(method reflect.Value, args ...interface{}) []reflect.Value {
	inputs := make([]reflect.Value, len(args))
	for i, _ := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}
	out := method.Call(inputs)
	return out
}

func (s *defaultServer) Invoke(handlerType interface{}, method string, ctx context.Context, message proto.Message) error {
	if handlerType == nil {
		e := fmt.Errorf("handlerType is nil")
		return e
	}
	found, exists := s.handleTypeMap[handlerType]
	ht := reflect.TypeOf(handlerType).Elem()
	if !exists {
		grpcMethods, e := parseService(ht)
		logger.Infof("parse service: %v to grpcMethods: %v", ht, grpcMethods)
		if e != nil {
			return e
		}
		found = grpcMethods
	}
	m, methodExists := found.MethodMap[method]
	if !methodExists {
		e := fmt.Errorf("could'nt found method: %v on type: %v. exists methods: %v", method, ht, found.MethodMap)
		return e
	}
	bytes, e := proto.Marshal(message)
	if e != nil {
		return e
	}
	if logger.IsDebugEnabled() {
		logger.Debugf("marshaled message: %v to: %v", message, bytes)
	}
	cmd := BuildUrl(ht, m.Method)
	mqttMessage := &mqttpb.QueueMessage{}
	mqttMessage.Command = cmd
	mqttMessage.Message = bytes
	raw, e := proto.Marshal(mqttMessage)
	if e != nil {
		return e
	}

	if logger.IsDebugEnabled() {
		logger.Debugf("marshaled mqttMessage: %v to: %v", mqttMessage, raw)
	}
	e = s.Produce(raw)
	return e
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
