// Code generated by protoc-gen-go. DO NOT EDIT.
// source: api.proto

package proto

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type MessageType int32

const (
	MessageType_REQUEST  MessageType = 0
	MessageType_RESPONSE MessageType = 1
)

var MessageType_name = map[int32]string{
	0: "REQUEST",
	1: "RESPONSE",
}

var MessageType_value = map[string]int32{
	"REQUEST":  0,
	"RESPONSE": 1,
}

func (x MessageType) String() string {
	return proto.EnumName(MessageType_name, int32(x))
}

func (MessageType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_00212fb1f9d3bf1c, []int{0}
}

type QueueMessage struct {
	MessageId            string      `protobuf:"bytes,3,opt,name=message_id,json=messageId,proto3" json:"message_id,omitempty"`
	Type                 MessageType `protobuf:"varint,4,opt,name=type,proto3,enum=proto.queue.MessageType" json:"type,omitempty"`
	Command              string      `protobuf:"bytes,1,opt,name=command,proto3" json:"command,omitempty"`
	Message              []byte      `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	Success              bool        `protobuf:"varint,5,opt,name=success,proto3" json:"success,omitempty"`
	Error                string      `protobuf:"bytes,6,opt,name=error,proto3" json:"error,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *QueueMessage) Reset()         { *m = QueueMessage{} }
func (m *QueueMessage) String() string { return proto.CompactTextString(m) }
func (*QueueMessage) ProtoMessage()    {}
func (*QueueMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_00212fb1f9d3bf1c, []int{0}
}

func (m *QueueMessage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_QueueMessage.Unmarshal(m, b)
}
func (m *QueueMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_QueueMessage.Marshal(b, m, deterministic)
}
func (m *QueueMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueueMessage.Merge(m, src)
}
func (m *QueueMessage) XXX_Size() int {
	return xxx_messageInfo_QueueMessage.Size(m)
}
func (m *QueueMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_QueueMessage.DiscardUnknown(m)
}

var xxx_messageInfo_QueueMessage proto.InternalMessageInfo

func (m *QueueMessage) GetMessageId() string {
	if m != nil {
		return m.MessageId
	}
	return ""
}

func (m *QueueMessage) GetType() MessageType {
	if m != nil {
		return m.Type
	}
	return MessageType_REQUEST
}

func (m *QueueMessage) GetCommand() string {
	if m != nil {
		return m.Command
	}
	return ""
}

func (m *QueueMessage) GetMessage() []byte {
	if m != nil {
		return m.Message
	}
	return nil
}

func (m *QueueMessage) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func (m *QueueMessage) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func init() {
	proto.RegisterEnum("proto.queue.MessageType", MessageType_name, MessageType_value)
	proto.RegisterType((*QueueMessage)(nil), "proto.queue.QueueMessage")
}

func init() { proto.RegisterFile("api.proto", fileDescriptor_00212fb1f9d3bf1c) }

var fileDescriptor_00212fb1f9d3bf1c = []byte{
	// 217 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4c, 0x2c, 0xc8, 0xd4,
	0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x06, 0x53, 0x7a, 0x85, 0xa5, 0xa9, 0xa5, 0xa9, 0x4a,
	0xfb, 0x19, 0xb9, 0x78, 0x02, 0x41, 0x2c, 0xdf, 0xd4, 0xe2, 0xe2, 0xc4, 0xf4, 0x54, 0x21, 0x59,
	0x2e, 0xae, 0x5c, 0x08, 0x33, 0x3e, 0x33, 0x45, 0x82, 0x59, 0x81, 0x51, 0x83, 0x33, 0x88, 0x13,
	0x2a, 0xe2, 0x99, 0x22, 0xa4, 0xc3, 0xc5, 0x52, 0x52, 0x59, 0x90, 0x2a, 0xc1, 0xa2, 0xc0, 0xa8,
	0xc1, 0x67, 0x24, 0xa1, 0x87, 0x64, 0x96, 0x1e, 0xd4, 0x88, 0x90, 0xca, 0x82, 0xd4, 0x20, 0xb0,
	0x2a, 0x21, 0x09, 0x2e, 0xf6, 0xe4, 0xfc, 0xdc, 0xdc, 0xc4, 0xbc, 0x14, 0x09, 0x46, 0xb0, 0x49,
	0x30, 0x2e, 0x48, 0x06, 0x6a, 0xa8, 0x04, 0x93, 0x02, 0xa3, 0x06, 0x4f, 0x10, 0x8c, 0x0b, 0x92,
	0x29, 0x2e, 0x4d, 0x4e, 0x4e, 0x2d, 0x2e, 0x96, 0x60, 0x55, 0x60, 0xd4, 0xe0, 0x08, 0x82, 0x71,
	0x85, 0x44, 0xb8, 0x58, 0x53, 0x8b, 0x8a, 0xf2, 0x8b, 0x24, 0xd8, 0xc0, 0x66, 0x41, 0x38, 0x5a,
	0x1a, 0x5c, 0xdc, 0x48, 0x16, 0x0b, 0x71, 0x73, 0xb1, 0x07, 0xb9, 0x06, 0x86, 0xba, 0x06, 0x87,
	0x08, 0x30, 0x08, 0xf1, 0x70, 0x71, 0x04, 0xb9, 0x06, 0x07, 0xf8, 0xfb, 0x05, 0xbb, 0x0a, 0x30,
	0x3a, 0xb1, 0x47, 0xb1, 0x82, 0x9d, 0x9b, 0xc4, 0x06, 0xa6, 0x8c, 0x01, 0x01, 0x00, 0x00, 0xff,
	0xff, 0x46, 0xfb, 0xac, 0x23, 0x15, 0x01, 0x00, 0x00,
}
