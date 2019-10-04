// Code generated by protoc-gen-go. DO NOT EDIT.
// source: ping.proto

package srv

import (
	context "context"
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
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

type PingMessage struct {
	Seq                  int64    `protobuf:"varint,1,opt,name=seq,proto3" json:"seq,omitempty"`
	Ts                   int64    `protobuf:"varint,2,opt,name=ts,proto3" json:"ts,omitempty"`
	Payload              string   `protobuf:"bytes,3,opt,name=payload,proto3" json:"payload,omitempty"`
	DelayNanos           int64    `protobuf:"varint,4,opt,name=delayNanos,proto3" json:"delayNanos,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PingMessage) Reset()         { *m = PingMessage{} }
func (m *PingMessage) String() string { return proto.CompactTextString(m) }
func (*PingMessage) ProtoMessage()    {}
func (*PingMessage) Descriptor() ([]byte, []int) {
	return fileDescriptor_6d51d96c3ad891f5, []int{0}
}

func (m *PingMessage) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PingMessage.Unmarshal(m, b)
}
func (m *PingMessage) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PingMessage.Marshal(b, m, deterministic)
}
func (m *PingMessage) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PingMessage.Merge(m, src)
}
func (m *PingMessage) XXX_Size() int {
	return xxx_messageInfo_PingMessage.Size(m)
}
func (m *PingMessage) XXX_DiscardUnknown() {
	xxx_messageInfo_PingMessage.DiscardUnknown(m)
}

var xxx_messageInfo_PingMessage proto.InternalMessageInfo

func (m *PingMessage) GetSeq() int64 {
	if m != nil {
		return m.Seq
	}
	return 0
}

func (m *PingMessage) GetTs() int64 {
	if m != nil {
		return m.Ts
	}
	return 0
}

func (m *PingMessage) GetPayload() string {
	if m != nil {
		return m.Payload
	}
	return ""
}

func (m *PingMessage) GetDelayNanos() int64 {
	if m != nil {
		return m.DelayNanos
	}
	return 0
}

func init() {
	proto.RegisterType((*PingMessage)(nil), "srv.PingMessage")
}

func init() { proto.RegisterFile("ping.proto", fileDescriptor_6d51d96c3ad891f5) }

var fileDescriptor_6d51d96c3ad891f5 = []byte{
	// 167 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2a, 0xc8, 0xcc, 0x4b,
	0xd7, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x2e, 0x2e, 0x2a, 0x53, 0xca, 0xe4, 0xe2, 0x0e,
	0xc8, 0xcc, 0x4b, 0xf7, 0x4d, 0x2d, 0x2e, 0x4e, 0x4c, 0x4f, 0x15, 0x12, 0xe0, 0x62, 0x2e, 0x4e,
	0x2d, 0x94, 0x60, 0x54, 0x60, 0xd4, 0x60, 0x0e, 0x02, 0x31, 0x85, 0xf8, 0xb8, 0x98, 0x4a, 0x8a,
	0x25, 0x98, 0xc0, 0x02, 0x4c, 0x25, 0xc5, 0x42, 0x12, 0x5c, 0xec, 0x05, 0x89, 0x95, 0x39, 0xf9,
	0x89, 0x29, 0x12, 0xcc, 0x0a, 0x8c, 0x1a, 0x9c, 0x41, 0x30, 0xae, 0x90, 0x1c, 0x17, 0x57, 0x4a,
	0x6a, 0x4e, 0x62, 0xa5, 0x5f, 0x62, 0x5e, 0x7e, 0xb1, 0x04, 0x0b, 0x58, 0x07, 0x92, 0x88, 0x91,
	0x15, 0x17, 0x17, 0xc8, 0xaa, 0xe0, 0xd4, 0xa2, 0xb2, 0xd4, 0x22, 0x21, 0x1d, 0x2e, 0x16, 0x10,
	0x4f, 0x48, 0x40, 0xaf, 0xb8, 0xa8, 0x4c, 0x0f, 0xc9, 0x0d, 0x52, 0x18, 0x22, 0x4a, 0x0c, 0x4e,
	0xac, 0x51, 0x20, 0xd7, 0x26, 0xb1, 0x81, 0x5d, 0x6e, 0x0c, 0x08, 0x00, 0x00, 0xff, 0xff, 0x12,
	0x41, 0xcf, 0x24, 0xc7, 0x00, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// PingServerClient is the client API for PingServer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type PingServerClient interface {
	Ping(ctx context.Context, in *PingMessage, opts ...grpc.CallOption) (*PingMessage, error)
}

type pingServerClient struct {
	cc *grpc.ClientConn
}

func NewPingServerClient(cc *grpc.ClientConn) PingServerClient {
	return &pingServerClient{cc}
}

func (c *pingServerClient) Ping(ctx context.Context, in *PingMessage, opts ...grpc.CallOption) (*PingMessage, error) {
	out := new(PingMessage)
	err := c.cc.Invoke(ctx, "/srv.PingServer/Ping", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PingServerServer is the server API for PingServer service.
type PingServerServer interface {
	Ping(context.Context, *PingMessage) (*PingMessage, error)
}

// UnimplementedPingServerServer can be embedded to have forward compatible implementations.
type UnimplementedPingServerServer struct {
}

func (*UnimplementedPingServerServer) Ping(ctx context.Context, req *PingMessage) (*PingMessage, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}

func RegisterPingServerServer(s *grpc.Server, srv PingServerServer) {
	s.RegisterService(&_PingServer_serviceDesc, srv)
}

func _PingServer_Ping_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PingMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PingServerServer).Ping(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/srv.PingServer/Ping",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PingServerServer).Ping(ctx, req.(*PingMessage))
	}
	return interceptor(ctx, in, info, handler)
}

var _PingServer_serviceDesc = grpc.ServiceDesc{
	ServiceName: "srv.PingServer",
	HandlerType: (*PingServerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Ping",
			Handler:    _PingServer_Ping_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "ping.proto",
}