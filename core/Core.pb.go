// Code generated by protoc-gen-go. DO NOT EDIT.
// source: Core.proto

/*
Package opencopilot is a generated protocol buffer package.

It is generated from these files:
	Core.proto

It has these top-level messages:
	Auth
	ListInstancesRequest
	CreateInstanceRequest
	DestroyInstanceRequest
	DestroyInstanceResponse
	GetInstanceRequest
	AddServiceRequest
	RemoveServiceRequest
	Instance
	Device
	Service
*/
package opencopilot

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Provider int32

const (
	Provider_PACKET Provider = 0
)

var Provider_name = map[int32]string{
	0: "PACKET",
}
var Provider_value = map[string]int32{
	"PACKET": 0,
}

func (x Provider) String() string {
	return proto.EnumName(Provider_name, int32(x))
}
func (Provider) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type Auth struct {
	Provider Provider `protobuf:"varint,1,opt,name=provider,enum=opencopilot.Provider" json:"provider,omitempty"`
	Payload  string   `protobuf:"bytes,2,opt,name=payload" json:"payload,omitempty"`
}

func (m *Auth) Reset()                    { *m = Auth{} }
func (m *Auth) String() string            { return proto.CompactTextString(m) }
func (*Auth) ProtoMessage()               {}
func (*Auth) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Auth) GetProvider() Provider {
	if m != nil {
		return m.Provider
	}
	return Provider_PACKET
}

func (m *Auth) GetPayload() string {
	if m != nil {
		return m.Payload
	}
	return ""
}

type ListInstancesRequest struct {
	Auth *Auth `protobuf:"bytes,1,opt,name=auth" json:"auth,omitempty"`
}

func (m *ListInstancesRequest) Reset()                    { *m = ListInstancesRequest{} }
func (m *ListInstancesRequest) String() string            { return proto.CompactTextString(m) }
func (*ListInstancesRequest) ProtoMessage()               {}
func (*ListInstancesRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *ListInstancesRequest) GetAuth() *Auth {
	if m != nil {
		return m.Auth
	}
	return nil
}

type CreateInstanceRequest struct {
	Auth *Auth `protobuf:"bytes,1,opt,name=auth" json:"auth,omitempty"`
}

func (m *CreateInstanceRequest) Reset()                    { *m = CreateInstanceRequest{} }
func (m *CreateInstanceRequest) String() string            { return proto.CompactTextString(m) }
func (*CreateInstanceRequest) ProtoMessage()               {}
func (*CreateInstanceRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *CreateInstanceRequest) GetAuth() *Auth {
	if m != nil {
		return m.Auth
	}
	return nil
}

type DestroyInstanceRequest struct {
	Auth       *Auth  `protobuf:"bytes,1,opt,name=auth" json:"auth,omitempty"`
	InstanceId string `protobuf:"bytes,2,opt,name=instance_id,json=instanceId" json:"instance_id,omitempty"`
}

func (m *DestroyInstanceRequest) Reset()                    { *m = DestroyInstanceRequest{} }
func (m *DestroyInstanceRequest) String() string            { return proto.CompactTextString(m) }
func (*DestroyInstanceRequest) ProtoMessage()               {}
func (*DestroyInstanceRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *DestroyInstanceRequest) GetAuth() *Auth {
	if m != nil {
		return m.Auth
	}
	return nil
}

func (m *DestroyInstanceRequest) GetInstanceId() string {
	if m != nil {
		return m.InstanceId
	}
	return ""
}

type DestroyInstanceResponse struct {
}

func (m *DestroyInstanceResponse) Reset()                    { *m = DestroyInstanceResponse{} }
func (m *DestroyInstanceResponse) String() string            { return proto.CompactTextString(m) }
func (*DestroyInstanceResponse) ProtoMessage()               {}
func (*DestroyInstanceResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type GetInstanceRequest struct {
	Auth       *Auth  `protobuf:"bytes,1,opt,name=auth" json:"auth,omitempty"`
	InstanceId string `protobuf:"bytes,2,opt,name=instance_id,json=instanceId" json:"instance_id,omitempty"`
}

func (m *GetInstanceRequest) Reset()                    { *m = GetInstanceRequest{} }
func (m *GetInstanceRequest) String() string            { return proto.CompactTextString(m) }
func (*GetInstanceRequest) ProtoMessage()               {}
func (*GetInstanceRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *GetInstanceRequest) GetAuth() *Auth {
	if m != nil {
		return m.Auth
	}
	return nil
}

func (m *GetInstanceRequest) GetInstanceId() string {
	if m != nil {
		return m.InstanceId
	}
	return ""
}

type AddServiceRequest struct {
	Auth       *Auth  `protobuf:"bytes,1,opt,name=auth" json:"auth,omitempty"`
	InstanceId string `protobuf:"bytes,2,opt,name=instance_id,json=instanceId" json:"instance_id,omitempty"`
	Service    string `protobuf:"bytes,3,opt,name=service" json:"service,omitempty"`
	Config     string `protobuf:"bytes,4,opt,name=config" json:"config,omitempty"`
}

func (m *AddServiceRequest) Reset()                    { *m = AddServiceRequest{} }
func (m *AddServiceRequest) String() string            { return proto.CompactTextString(m) }
func (*AddServiceRequest) ProtoMessage()               {}
func (*AddServiceRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *AddServiceRequest) GetAuth() *Auth {
	if m != nil {
		return m.Auth
	}
	return nil
}

func (m *AddServiceRequest) GetInstanceId() string {
	if m != nil {
		return m.InstanceId
	}
	return ""
}

func (m *AddServiceRequest) GetService() string {
	if m != nil {
		return m.Service
	}
	return ""
}

func (m *AddServiceRequest) GetConfig() string {
	if m != nil {
		return m.Config
	}
	return ""
}

type RemoveServiceRequest struct {
	Auth       *Auth  `protobuf:"bytes,1,opt,name=auth" json:"auth,omitempty"`
	InstanceId string `protobuf:"bytes,2,opt,name=instance_id,json=instanceId" json:"instance_id,omitempty"`
	Service    string `protobuf:"bytes,3,opt,name=service" json:"service,omitempty"`
}

func (m *RemoveServiceRequest) Reset()                    { *m = RemoveServiceRequest{} }
func (m *RemoveServiceRequest) String() string            { return proto.CompactTextString(m) }
func (*RemoveServiceRequest) ProtoMessage()               {}
func (*RemoveServiceRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *RemoveServiceRequest) GetAuth() *Auth {
	if m != nil {
		return m.Auth
	}
	return nil
}

func (m *RemoveServiceRequest) GetInstanceId() string {
	if m != nil {
		return m.InstanceId
	}
	return ""
}

func (m *RemoveServiceRequest) GetService() string {
	if m != nil {
		return m.Service
	}
	return ""
}

type Instance struct {
	Id       string     `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Provider Provider   `protobuf:"varint,2,opt,name=provider,enum=opencopilot.Provider" json:"provider,omitempty"`
	Owner    string     `protobuf:"bytes,3,opt,name=owner" json:"owner,omitempty"`
	Device   string     `protobuf:"bytes,4,opt,name=device" json:"device,omitempty"`
	Services []*Service `protobuf:"bytes,5,rep,name=services" json:"services,omitempty"`
}

func (m *Instance) Reset()                    { *m = Instance{} }
func (m *Instance) String() string            { return proto.CompactTextString(m) }
func (*Instance) ProtoMessage()               {}
func (*Instance) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *Instance) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Instance) GetProvider() Provider {
	if m != nil {
		return m.Provider
	}
	return Provider_PACKET
}

func (m *Instance) GetOwner() string {
	if m != nil {
		return m.Owner
	}
	return ""
}

func (m *Instance) GetDevice() string {
	if m != nil {
		return m.Device
	}
	return ""
}

func (m *Instance) GetServices() []*Service {
	if m != nil {
		return m.Services
	}
	return nil
}

type Device struct {
	Id     string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Status string `protobuf:"bytes,2,opt,name=status" json:"status,omitempty"`
}

func (m *Device) Reset()                    { *m = Device{} }
func (m *Device) String() string            { return proto.CompactTextString(m) }
func (*Device) ProtoMessage()               {}
func (*Device) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *Device) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Device) GetStatus() string {
	if m != nil {
		return m.Status
	}
	return ""
}

type Service struct {
	Type   string `protobuf:"bytes,1,opt,name=type" json:"type,omitempty"`
	Config string `protobuf:"bytes,2,opt,name=config" json:"config,omitempty"`
}

func (m *Service) Reset()                    { *m = Service{} }
func (m *Service) String() string            { return proto.CompactTextString(m) }
func (*Service) ProtoMessage()               {}
func (*Service) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func (m *Service) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *Service) GetConfig() string {
	if m != nil {
		return m.Config
	}
	return ""
}

func init() {
	proto.RegisterType((*Auth)(nil), "opencopilot.Auth")
	proto.RegisterType((*ListInstancesRequest)(nil), "opencopilot.ListInstancesRequest")
	proto.RegisterType((*CreateInstanceRequest)(nil), "opencopilot.CreateInstanceRequest")
	proto.RegisterType((*DestroyInstanceRequest)(nil), "opencopilot.DestroyInstanceRequest")
	proto.RegisterType((*DestroyInstanceResponse)(nil), "opencopilot.DestroyInstanceResponse")
	proto.RegisterType((*GetInstanceRequest)(nil), "opencopilot.GetInstanceRequest")
	proto.RegisterType((*AddServiceRequest)(nil), "opencopilot.AddServiceRequest")
	proto.RegisterType((*RemoveServiceRequest)(nil), "opencopilot.RemoveServiceRequest")
	proto.RegisterType((*Instance)(nil), "opencopilot.Instance")
	proto.RegisterType((*Device)(nil), "opencopilot.Device")
	proto.RegisterType((*Service)(nil), "opencopilot.Service")
	proto.RegisterEnum("opencopilot.Provider", Provider_name, Provider_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Core service

type CoreClient interface {
	ListInstances(ctx context.Context, in *ListInstancesRequest, opts ...grpc.CallOption) (Core_ListInstancesClient, error)
	CreateInstance(ctx context.Context, in *CreateInstanceRequest, opts ...grpc.CallOption) (*Instance, error)
	DestroyInstance(ctx context.Context, in *DestroyInstanceRequest, opts ...grpc.CallOption) (*DestroyInstanceResponse, error)
	GetInstance(ctx context.Context, in *GetInstanceRequest, opts ...grpc.CallOption) (*Instance, error)
	AddService(ctx context.Context, in *AddServiceRequest, opts ...grpc.CallOption) (*Instance, error)
	RemoveService(ctx context.Context, in *RemoveServiceRequest, opts ...grpc.CallOption) (*Instance, error)
}

type coreClient struct {
	cc *grpc.ClientConn
}

func NewCoreClient(cc *grpc.ClientConn) CoreClient {
	return &coreClient{cc}
}

func (c *coreClient) ListInstances(ctx context.Context, in *ListInstancesRequest, opts ...grpc.CallOption) (Core_ListInstancesClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Core_serviceDesc.Streams[0], c.cc, "/opencopilot.Core/ListInstances", opts...)
	if err != nil {
		return nil, err
	}
	x := &coreListInstancesClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Core_ListInstancesClient interface {
	Recv() (*Instance, error)
	grpc.ClientStream
}

type coreListInstancesClient struct {
	grpc.ClientStream
}

func (x *coreListInstancesClient) Recv() (*Instance, error) {
	m := new(Instance)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *coreClient) CreateInstance(ctx context.Context, in *CreateInstanceRequest, opts ...grpc.CallOption) (*Instance, error) {
	out := new(Instance)
	err := grpc.Invoke(ctx, "/opencopilot.Core/CreateInstance", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *coreClient) DestroyInstance(ctx context.Context, in *DestroyInstanceRequest, opts ...grpc.CallOption) (*DestroyInstanceResponse, error) {
	out := new(DestroyInstanceResponse)
	err := grpc.Invoke(ctx, "/opencopilot.Core/DestroyInstance", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *coreClient) GetInstance(ctx context.Context, in *GetInstanceRequest, opts ...grpc.CallOption) (*Instance, error) {
	out := new(Instance)
	err := grpc.Invoke(ctx, "/opencopilot.Core/GetInstance", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *coreClient) AddService(ctx context.Context, in *AddServiceRequest, opts ...grpc.CallOption) (*Instance, error) {
	out := new(Instance)
	err := grpc.Invoke(ctx, "/opencopilot.Core/AddService", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *coreClient) RemoveService(ctx context.Context, in *RemoveServiceRequest, opts ...grpc.CallOption) (*Instance, error) {
	out := new(Instance)
	err := grpc.Invoke(ctx, "/opencopilot.Core/RemoveService", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Core service

type CoreServer interface {
	ListInstances(*ListInstancesRequest, Core_ListInstancesServer) error
	CreateInstance(context.Context, *CreateInstanceRequest) (*Instance, error)
	DestroyInstance(context.Context, *DestroyInstanceRequest) (*DestroyInstanceResponse, error)
	GetInstance(context.Context, *GetInstanceRequest) (*Instance, error)
	AddService(context.Context, *AddServiceRequest) (*Instance, error)
	RemoveService(context.Context, *RemoveServiceRequest) (*Instance, error)
}

func RegisterCoreServer(s *grpc.Server, srv CoreServer) {
	s.RegisterService(&_Core_serviceDesc, srv)
}

func _Core_ListInstances_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(ListInstancesRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(CoreServer).ListInstances(m, &coreListInstancesServer{stream})
}

type Core_ListInstancesServer interface {
	Send(*Instance) error
	grpc.ServerStream
}

type coreListInstancesServer struct {
	grpc.ServerStream
}

func (x *coreListInstancesServer) Send(m *Instance) error {
	return x.ServerStream.SendMsg(m)
}

func _Core_CreateInstance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateInstanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoreServer).CreateInstance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/opencopilot.Core/CreateInstance",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoreServer).CreateInstance(ctx, req.(*CreateInstanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Core_DestroyInstance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DestroyInstanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoreServer).DestroyInstance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/opencopilot.Core/DestroyInstance",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoreServer).DestroyInstance(ctx, req.(*DestroyInstanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Core_GetInstance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetInstanceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoreServer).GetInstance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/opencopilot.Core/GetInstance",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoreServer).GetInstance(ctx, req.(*GetInstanceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Core_AddService_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddServiceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoreServer).AddService(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/opencopilot.Core/AddService",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoreServer).AddService(ctx, req.(*AddServiceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Core_RemoveService_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveServiceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CoreServer).RemoveService(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/opencopilot.Core/RemoveService",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CoreServer).RemoveService(ctx, req.(*RemoveServiceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Core_serviceDesc = grpc.ServiceDesc{
	ServiceName: "opencopilot.Core",
	HandlerType: (*CoreServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateInstance",
			Handler:    _Core_CreateInstance_Handler,
		},
		{
			MethodName: "DestroyInstance",
			Handler:    _Core_DestroyInstance_Handler,
		},
		{
			MethodName: "GetInstance",
			Handler:    _Core_GetInstance_Handler,
		},
		{
			MethodName: "AddService",
			Handler:    _Core_AddService_Handler,
		},
		{
			MethodName: "RemoveService",
			Handler:    _Core_RemoveService_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "ListInstances",
			Handler:       _Core_ListInstances_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "Core.proto",
}

func init() { proto.RegisterFile("Core.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 495 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xbc, 0x54, 0xc1, 0x6e, 0xd3, 0x40,
	0x10, 0x8d, 0x1d, 0xd7, 0x4d, 0x9f, 0xd5, 0x40, 0x47, 0x69, 0x30, 0x3d, 0xd0, 0xb0, 0x80, 0x14,
	0x71, 0x88, 0x42, 0x10, 0x47, 0x90, 0xa2, 0xb4, 0xaa, 0xaa, 0x52, 0xa9, 0x72, 0x39, 0x22, 0xc0,
	0xc4, 0x0b, 0x5d, 0xa9, 0x78, 0x8d, 0x77, 0x13, 0xc8, 0x47, 0xf0, 0x29, 0x7c, 0x14, 0x7f, 0x82,
	0x62, 0xaf, 0x4b, 0x36, 0x4d, 0x10, 0x20, 0xe0, 0xe6, 0xf1, 0xcc, 0xbc, 0xf7, 0x66, 0x67, 0xdf,
	0x02, 0x23, 0x99, 0xf3, 0x5e, 0x96, 0x4b, 0x2d, 0x29, 0x90, 0x19, 0x4f, 0xc7, 0x32, 0x13, 0x97,
	0x52, 0xb3, 0x73, 0x78, 0xc3, 0x89, 0xbe, 0xa0, 0x47, 0x68, 0x64, 0xb9, 0x9c, 0x8a, 0x84, 0xe7,
	0xa1, 0xd3, 0x71, 0xba, 0xcd, 0xc1, 0x6e, 0x6f, 0xa1, 0xae, 0x77, 0x66, 0x92, 0xd1, 0x55, 0x19,
	0x85, 0xd8, 0xcc, 0xe2, 0xd9, 0xa5, 0x8c, 0x93, 0xd0, 0xed, 0x38, 0xdd, 0xad, 0xa8, 0x0a, 0xd9,
	0x53, 0xb4, 0x9e, 0x0b, 0xa5, 0x8f, 0x53, 0xa5, 0xe3, 0x74, 0xcc, 0x55, 0xc4, 0x3f, 0x4e, 0xb8,
	0xd2, 0xf4, 0x00, 0x5e, 0x3c, 0xd1, 0x17, 0x05, 0x41, 0x30, 0xd8, 0xb1, 0x08, 0xe6, 0x2a, 0xa2,
	0x22, 0xcd, 0x9e, 0x61, 0x77, 0x94, 0xf3, 0x58, 0xf3, 0x0a, 0xe0, 0x37, 0xfb, 0xdf, 0xa0, 0x7d,
	0xc0, 0x95, 0xce, 0xe5, 0xec, 0xcf, 0x00, 0x68, 0x1f, 0x81, 0x30, 0x9d, 0xaf, 0x45, 0x35, 0x1d,
	0xaa, 0x5f, 0xc7, 0x09, 0xbb, 0x8d, 0x5b, 0xd7, 0x18, 0x54, 0x26, 0x53, 0xc5, 0xd9, 0x4b, 0xd0,
	0x11, 0xd7, 0xff, 0x8a, 0xf8, 0x8b, 0x83, 0x9d, 0x61, 0x92, 0x9c, 0xf3, 0x7c, 0x2a, 0xfe, 0x3a,
	0xfa, 0x7c, 0xa3, 0xaa, 0x44, 0x0e, 0xeb, 0xe5, 0x46, 0x4d, 0x48, 0x6d, 0xf8, 0x63, 0x99, 0xbe,
	0x13, 0xef, 0x43, 0xaf, 0x48, 0x98, 0x88, 0x7d, 0x46, 0x2b, 0xe2, 0x1f, 0xe4, 0x94, 0xff, 0x6f,
	0x45, 0xec, 0xab, 0x83, 0x46, 0x75, 0xca, 0xd4, 0x84, 0x2b, 0x92, 0x82, 0x6c, 0x2b, 0x72, 0x45,
	0x62, 0xdd, 0x66, 0xf7, 0xd7, 0x6e, 0x73, 0x0b, 0x1b, 0xf2, 0x53, 0xca, 0x73, 0xc3, 0x53, 0x06,
	0xf3, 0xb9, 0x13, 0x5e, 0xd0, 0x9b, 0xb9, 0xcb, 0x88, 0xfa, 0x68, 0x18, 0x21, 0x2a, 0xdc, 0xe8,
	0xd4, 0xbb, 0xc1, 0xa0, 0x65, 0x11, 0x54, 0xc7, 0x71, 0x55, 0xc5, 0xfa, 0xf0, 0x0f, 0xca, 0xde,
	0x65, 0xb1, 0x6d, 0xf8, 0x4a, 0xc7, 0x7a, 0xa2, 0xcc, 0xfc, 0x26, 0x62, 0x4f, 0xb0, 0x69, 0x60,
	0x88, 0xe0, 0xe9, 0x59, 0xc6, 0x4d, 0x53, 0xf1, 0xbd, 0xb0, 0x12, 0x77, 0x71, 0x25, 0x0f, 0xdb,
	0x68, 0x54, 0xe3, 0x11, 0xe0, 0x9f, 0x0d, 0x47, 0x27, 0x87, 0x2f, 0x6e, 0xd6, 0x06, 0xdf, 0xea,
	0xf0, 0xe6, 0xaf, 0x00, 0x9d, 0x62, 0xdb, 0x72, 0x27, 0xdd, 0xb5, 0xa4, 0xaf, 0x72, 0xee, 0x9e,
	0x7d, 0x7c, 0x55, 0x9a, 0xd5, 0xfa, 0x0e, 0x9d, 0xa2, 0x69, 0xbb, 0x95, 0x98, 0x55, 0xbc, 0xd2,
	0xca, 0x6b, 0x01, 0xe9, 0x15, 0x6e, 0x2c, 0x59, 0x8b, 0xee, 0x59, 0xb5, 0xab, 0xad, 0xbd, 0x77,
	0xff, 0xe7, 0x45, 0xc6, 0x9d, 0x35, 0x3a, 0x42, 0xb0, 0xe0, 0x4f, 0xda, 0xb7, 0xda, 0xae, 0x3b,
	0x77, 0xbd, 0xd0, 0x43, 0xe0, 0x87, 0x13, 0xe9, 0x8e, 0x7d, 0xc5, 0x97, 0x2d, 0xba, 0x1e, 0xe6,
	0x04, 0xdb, 0x96, 0x83, 0x96, 0xb6, 0xb1, 0xca, 0x5d, 0x6b, 0xc1, 0xde, 0xfa, 0xc5, 0x0b, 0xff,
	0xf8, 0x7b, 0x00, 0x00, 0x00, 0xff, 0xff, 0x93, 0xd2, 0x8b, 0xa7, 0xef, 0x05, 0x00, 0x00,
}
