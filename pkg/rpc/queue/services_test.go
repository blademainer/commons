package queue

import (
	"context"
	"fmt"
	"github.com/blademainer/commons/pkg/logger"
	mqttpb "github.com/blademainer/commons/pkg/rpc/queue/proto"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gopkg.in/go-playground/assert.v1"
	"math"
	"reflect"
	"testing"
)

// server is used to implement helloworld.GreeterServer.
type server struct{}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *HelloRequest) (*HelloReply, error) {
	return &HelloReply{Message: "Hello " + in.Name}, nil
}

func BenchmarkHandle(b *testing.B) {
	logger.SetLevel(logger.LOG_LEVEL_INFO)

	controller := gomock.NewController(b)
	queue := NewMockQueue(controller)
	var capturedArgs [][]byte
	queue.
		EXPECT().
		Produce(gomock.Any()).
		Do(func(arg []byte) {
			capturedArgs = append(capturedArgs, arg)
		}).MinTimes(1)
	//queue.EXPECT().Produce(gomock.Any()).Return(nil).Times(1)

	opts := NewOptions().InvokeTimeout(5 * time.Second)
	s := NewServer(queue, opts)
	handlerType := (*GreeterServer)(nil)

	svc := &server{}
	e := s.RegisterService(handlerType, svc)
	if e != nil {
		logger.Fatal(e)
	}

	request := &HelloRequest{Name: "zhangsan"}
	respose, e := s.Invoke(handlerType, "SayHello", context.Background(), request)
	fmt.Println(respose)
	for i := 0; i < b.N; i++ {
		e = s.Handle(capturedArgs[0])
		if e != nil {
			logger.Fatal(e)
		}
	}
}

func Test_defaultServer_parseService(t *testing.T) {
	handlerType := (*GreeterServer)(nil)
	ht := reflect.TypeOf(handlerType).Elem()
	methods, e := parseService(ht)
	if e != nil {
		panic(e)
	}
	for _, m := range methods.Methods {
		if m.In == reflect.TypeOf((*HelloRequest)(nil)) {
			name := "hello tests"
			request := &HelloRequest{Name: name}
			bytes, e := proto.Marshal(request)
			if e != nil {
				t.Fatal(e)
			}
			messageType := reflect.TypeOf((*proto.Message)(nil)).Elem()
			if !m.In.Implements(messageType) {
				panic("not implements *proto.Message")
			}
			value := reflect.New(m.In.Elem())
			fmt.Println(m.In, value.Kind())
			message := value.Interface().(proto.Message)
			e = proto.Unmarshal(bytes, message)
			if e != nil {
				t.Fatal(e)
			}
			fmt.Println(m)
			assert.Equal(t, name, message.(*HelloRequest).Name)
			//m.Method.Func.Call([]reflect.Value{context.Background(), message})
			// invokeRequest grpc method
			invoke := Invoke(&server{}, m.Method.Name, context.Background(), message)
			fmt.Println(invoke[0])

			marshal, e := proto.Marshal(invoke[0].Interface().(proto.Message))
			fmt.Println("marshal: ", marshal)
			fmt.Println("invokeRequest out: ", invoke)
		}
	}
	fmt.Println(methods)
}

func Test_defaultServer_Invoke(t *testing.T) {
	logger.SetLevel(logger.LOG_LEVEL_DEBUG)

	controller := gomock.NewController(t)
	queue := NewMockQueue(controller)
	var capturedArgs []byte

	queue.
		EXPECT().
		Produce(gomock.Any()).
		Do(func(arg []byte) {
			capturedArgs = arg
		}).Times(1)
	//queue.EXPECT().Produce(gomock.Any()).Return(nil).Times(1)

	opts := NewOptions().InvokeTimeout(5 * time.Second)
	s := NewServer(queue, opts)
	handlerType := (*GreeterServer)(nil)
	//s := &defaultServer{}
	//s.RegisterService(handlerType, svc)
	request := &HelloRequest{Name: "zhangsan"}
	response, e := s.Invoke(handlerType, "SayHello", context.Background(), request)
	fmt.Println(response)
	fmt.Println(capturedArgs)
	msg := &mqttpb.QueueMessage{}
	e = proto.Unmarshal(capturedArgs, msg)
	if e != nil {
		t.Fatalf(e.Error())
	}
	fmt.Println("marshal to msg: ", msg)
	assert.Equal(t, msg.Command, "/queue.GreeterServer/SayHello")

	helloRequest := &HelloRequest{}
	bytes := msg.Message
	e = proto.Unmarshal(bytes, helloRequest)
	if e != nil {
		t.Fatalf(e.Error())
	}
	assert.Equal(t, helloRequest.Name, request.Name)
}

func Test_defaultServer_RegisterService(t *testing.T) {
	logger.SetLevel(logger.LOG_LEVEL_DEBUG)

	controller := gomock.NewController(t)
	queue := NewMockQueue(controller)
	var capturedArgs []byte
	queue.
		EXPECT().
		Produce(gomock.Any()).
		Do(func(arg []byte) {
			capturedArgs = arg
		}).Times(1)
	//queue.EXPECT().Produce(gomock.Any()).Return(nil).Times(1)

	opts := NewOptions().InvokeTimeout(5 * time.Second)
	s := NewServer(queue, opts)
	handlerType := (*GreeterServer)(nil)

	svc := &server{}
	e := s.RegisterService(handlerType, svc)
	if e != nil {
		t.Fatal(e)
	}

	request := &HelloRequest{Name: "zhangsan"}
	response, e := s.Invoke(handlerType, "SayHello", context.Background(), request)
	fmt.Println(response)
	fmt.Println(capturedArgs)
	msg := &mqttpb.QueueMessage{}
	e = proto.Unmarshal(capturedArgs, msg)
	if e != nil {
		t.Fatalf(e.Error())
	}
	fmt.Println("marshal to msg: ", msg)
	assert.Equal(t, msg.Command, "/queue.GreeterServer/SayHello")

	helloRequest := &HelloRequest{}
	bytes := msg.Message
	e = proto.Unmarshal(bytes, helloRequest)
	if e != nil {
		t.Fatalf(e.Error())
	}
	assert.Equal(t, helloRequest.Name, request.Name)
}
func Test_defaultServer_Handle(t *testing.T) {
	logger.SetLevel(logger.LOG_LEVEL_DEBUG)

	controller := gomock.NewController(t)
	queue := NewMockQueue(controller)
	var capturedArgs [][]byte
	queue.
		EXPECT().
		Produce(gomock.Any()).
		Do(func(arg []byte) {
			capturedArgs = append(capturedArgs, arg)
		}).MinTimes(1)
	//queue.EXPECT().Produce(gomock.Any()).Return(nil).Times(1)

	opts := NewOptions().InvokeTimeout(5 * time.Second)
	s := NewServer(queue, opts)
	handlerType := (*GreeterServer)(nil)

	svc := &server{}
	e := s.RegisterService(handlerType, svc)
	if e != nil {
		t.Fatal(e)
	}

	request := &HelloRequest{Name: "zhangsan"}
	response, e := s.Invoke(handlerType, "SayHello", context.Background(), request)
	fmt.Println(response)
	e = s.Handle(capturedArgs[0])
	if e != nil {
		t.Fatal(e)
	}

}

func Invoke(any interface{}, name string, args ...interface{}) []reflect.Value {
	inputs := make([]reflect.Value, len(args))
	for i, _ := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}
	out := reflect.ValueOf(any).MethodByName(name).Call(inputs)
	return out
}

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// The request message containing the user's name.
type HelloRequest struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HelloRequest) Reset()         { *m = HelloRequest{} }
func (m *HelloRequest) String() string { return proto.CompactTextString(m) }
func (*HelloRequest) ProtoMessage()    {}
func (*HelloRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_17b8c58d586b62f2, []int{0}
}

func (m *HelloRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HelloRequest.Unmarshal(m, b)
}
func (m *HelloRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HelloRequest.Marshal(b, m, deterministic)
}
func (m *HelloRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HelloRequest.Merge(m, src)
}
func (m *HelloRequest) XXX_Size() int {
	return xxx_messageInfo_HelloRequest.Size(m)
}
func (m *HelloRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_HelloRequest.DiscardUnknown(m)
}

var xxx_messageInfo_HelloRequest proto.InternalMessageInfo

func (m *HelloRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

// The response message containing the greetings
type HelloReply struct {
	Message              string   `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *HelloReply) Reset()         { *m = HelloReply{} }
func (m *HelloReply) String() string { return proto.CompactTextString(m) }
func (*HelloReply) ProtoMessage()    {}
func (*HelloReply) Descriptor() ([]byte, []int) {
	return fileDescriptor_17b8c58d586b62f2, []int{1}
}

func (m *HelloReply) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HelloReply.Unmarshal(m, b)
}
func (m *HelloReply) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HelloReply.Marshal(b, m, deterministic)
}
func (m *HelloReply) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HelloReply.Merge(m, src)
}
func (m *HelloReply) XXX_Size() int {
	return xxx_messageInfo_HelloReply.Size(m)
}
func (m *HelloReply) XXX_DiscardUnknown() {
	xxx_messageInfo_HelloReply.DiscardUnknown(m)
}

var xxx_messageInfo_HelloReply proto.InternalMessageInfo

func (m *HelloReply) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func init() {
	proto.RegisterType((*HelloRequest)(nil), "helloworld.HelloRequest")
	proto.RegisterType((*HelloReply)(nil), "helloworld.HelloReply")
}

func init() { proto.RegisterFile("helloworld.proto", fileDescriptor_17b8c58d586b62f2) }

var fileDescriptor_17b8c58d586b62f2 = []byte{
	// 175 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x12, 0xc8, 0x48, 0xcd, 0xc9,
	0xc9, 0x2f, 0xcf, 0x2f, 0xca, 0x49, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x42, 0x88,
	0x28, 0x29, 0x71, 0xf1, 0x78, 0x80, 0x78, 0x41, 0xa9, 0x85, 0xa5, 0xa9, 0xc5, 0x25, 0x42, 0x42,
	0x5c, 0x2c, 0x79, 0x89, 0xb9, 0xa9, 0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0x9c, 0x41, 0x60, 0xb6, 0x92,
	0x1a, 0x17, 0x17, 0x54, 0x4d, 0x41, 0x4e, 0xa5, 0x90, 0x04, 0x17, 0x7b, 0x6e, 0x6a, 0x71, 0x71,
	0x62, 0x3a, 0x4c, 0x11, 0x8c, 0x6b, 0xe4, 0xc9, 0xc5, 0xee, 0x5e, 0x94, 0x9a, 0x5a, 0x92, 0x5a,
	0x24, 0x64, 0xc7, 0xc5, 0x11, 0x9c, 0x58, 0x09, 0xd6, 0x25, 0x24, 0xa1, 0x87, 0xe4, 0x02, 0x64,
	0xcb, 0xa4, 0xc4, 0xb0, 0xc8, 0x14, 0xe4, 0x54, 0x2a, 0x31, 0x38, 0x19, 0x70, 0x49, 0x67, 0xe6,
	0xeb, 0xa5, 0x17, 0x15, 0x24, 0xeb, 0xa5, 0x56, 0x24, 0xe6, 0x16, 0xe4, 0xa4, 0x16, 0x23, 0xa9,
	0x75, 0xe2, 0x07, 0x2b, 0x0e, 0x07, 0xb1, 0x03, 0x40, 0x5e, 0x0a, 0x60, 0x4c, 0x62, 0x03, 0xfb,
	0xcd, 0x18, 0x10, 0x00, 0x00, 0xff, 0xff, 0x0f, 0xb7, 0xcd, 0xf2, 0xef, 0x00, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// GreeterClient is the queue API for Greeter service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type GreeterClient interface {
	// Sends a greeting
	SayHello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloReply, error)
}

type greeterClient struct {
	cc *grpc.ClientConn
}

func NewGreeterClient(cc *grpc.ClientConn) GreeterClient {
	return &greeterClient{cc}
}

func (c *greeterClient) SayHello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloReply, error) {
	out := new(HelloReply)
	err := c.cc.Invoke(ctx, "/helloworld.Greeter/SayHello", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// GreeterServer is the server API for Greeter service.
type GreeterServer interface {
	// Sends a greeting
	SayHello(context.Context, *HelloRequest) (*HelloReply, error)
}

// UnimplementedGreeterServer can be embedded to have forward compatible implementations.
type UnimplementedGreeterServer struct {
}

func (*UnimplementedGreeterServer) SayHello(ctx context.Context, req *HelloRequest) (*HelloReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SayHello not implemented")
}

func RegisterGreeterServer(s *grpc.Server, srv GreeterServer) {
	s.RegisterService(&_Greeter_serviceDesc, srv)
}

func _Greeter_SayHello_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HelloRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GreeterServer).SayHello(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/helloworld.Greeter/SayHello",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GreeterServer).SayHello(ctx, req.(*HelloRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Greeter_serviceDesc = grpc.ServiceDesc{
	ServiceName: "helloworld.Greeter",
	HandlerType: (*GreeterServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SayHello",
			Handler:    _Greeter_SayHello_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "helloworld.proto",
}
