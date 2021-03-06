// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package service

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// MafiaClient is the client API for Mafia service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MafiaClient interface {
	// Sends a greeting
	AddUser(ctx context.Context, in *AddUserRequest, opts ...grpc.CallOption) (*Response, error)
	DeleteUser(ctx context.Context, in *BaseUserRequest, opts ...grpc.CallOption) (*Response, error)
	GetUsers(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*GetUsersResponse, error)
	AccuseUser(ctx context.Context, in *AccuseUserRequest, opts ...grpc.CallOption) (*Response, error)
	VoteFinishDay(ctx context.Context, in *BaseUserRequest, opts ...grpc.CallOption) (*Response, error)
	InitCommunicationChannel(ctx context.Context, opts ...grpc.CallOption) (Mafia_InitCommunicationChannelClient, error)
}

type mafiaClient struct {
	cc grpc.ClientConnInterface
}

func NewMafiaClient(cc grpc.ClientConnInterface) MafiaClient {
	return &mafiaClient{cc}
}

func (c *mafiaClient) AddUser(ctx context.Context, in *AddUserRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/mafia.Mafia/add_user", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mafiaClient) DeleteUser(ctx context.Context, in *BaseUserRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/mafia.Mafia/delete_user", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mafiaClient) GetUsers(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*GetUsersResponse, error) {
	out := new(GetUsersResponse)
	err := c.cc.Invoke(ctx, "/mafia.Mafia/get_users", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mafiaClient) AccuseUser(ctx context.Context, in *AccuseUserRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/mafia.Mafia/accuse_user", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mafiaClient) VoteFinishDay(ctx context.Context, in *BaseUserRequest, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/mafia.Mafia/vote_finish_day", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mafiaClient) InitCommunicationChannel(ctx context.Context, opts ...grpc.CallOption) (Mafia_InitCommunicationChannelClient, error) {
	stream, err := c.cc.NewStream(ctx, &Mafia_ServiceDesc.Streams[0], "/mafia.Mafia/init_communication_channel", opts...)
	if err != nil {
		return nil, err
	}
	x := &mafiaInitCommunicationChannelClient{stream}
	return x, nil
}

type Mafia_InitCommunicationChannelClient interface {
	Send(*CommunicationRequest) error
	Recv() (*Response, error)
	grpc.ClientStream
}

type mafiaInitCommunicationChannelClient struct {
	grpc.ClientStream
}

func (x *mafiaInitCommunicationChannelClient) Send(m *CommunicationRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *mafiaInitCommunicationChannelClient) Recv() (*Response, error) {
	m := new(Response)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// MafiaServer is the server API for Mafia service.
// All implementations must embed UnimplementedMafiaServer
// for forward compatibility
type MafiaServer interface {
	// Sends a greeting
	AddUser(context.Context, *AddUserRequest) (*Response, error)
	DeleteUser(context.Context, *BaseUserRequest) (*Response, error)
	GetUsers(context.Context, *Empty) (*GetUsersResponse, error)
	AccuseUser(context.Context, *AccuseUserRequest) (*Response, error)
	VoteFinishDay(context.Context, *BaseUserRequest) (*Response, error)
	InitCommunicationChannel(Mafia_InitCommunicationChannelServer) error
	mustEmbedUnimplementedMafiaServer()
}

// UnimplementedMafiaServer must be embedded to have forward compatible implementations.
type UnimplementedMafiaServer struct {
}

func (UnimplementedMafiaServer) AddUser(context.Context, *AddUserRequest) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddUser not implemented")
}
func (UnimplementedMafiaServer) DeleteUser(context.Context, *BaseUserRequest) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUser not implemented")
}
func (UnimplementedMafiaServer) GetUsers(context.Context, *Empty) (*GetUsersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetUsers not implemented")
}
func (UnimplementedMafiaServer) AccuseUser(context.Context, *AccuseUserRequest) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AccuseUser not implemented")
}
func (UnimplementedMafiaServer) VoteFinishDay(context.Context, *BaseUserRequest) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VoteFinishDay not implemented")
}
func (UnimplementedMafiaServer) InitCommunicationChannel(Mafia_InitCommunicationChannelServer) error {
	return status.Errorf(codes.Unimplemented, "method InitCommunicationChannel not implemented")
}
func (UnimplementedMafiaServer) mustEmbedUnimplementedMafiaServer() {}

// UnsafeMafiaServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MafiaServer will
// result in compilation errors.
type UnsafeMafiaServer interface {
	mustEmbedUnimplementedMafiaServer()
}

func RegisterMafiaServer(s grpc.ServiceRegistrar, srv MafiaServer) {
	s.RegisterService(&Mafia_ServiceDesc, srv)
}

func _Mafia_AddUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MafiaServer).AddUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mafia.Mafia/add_user",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MafiaServer).AddUser(ctx, req.(*AddUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Mafia_DeleteUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BaseUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MafiaServer).DeleteUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mafia.Mafia/delete_user",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MafiaServer).DeleteUser(ctx, req.(*BaseUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Mafia_GetUsers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MafiaServer).GetUsers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mafia.Mafia/get_users",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MafiaServer).GetUsers(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Mafia_AccuseUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AccuseUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MafiaServer).AccuseUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mafia.Mafia/accuse_user",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MafiaServer).AccuseUser(ctx, req.(*AccuseUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Mafia_VoteFinishDay_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BaseUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MafiaServer).VoteFinishDay(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mafia.Mafia/vote_finish_day",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MafiaServer).VoteFinishDay(ctx, req.(*BaseUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Mafia_InitCommunicationChannel_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(MafiaServer).InitCommunicationChannel(&mafiaInitCommunicationChannelServer{stream})
}

type Mafia_InitCommunicationChannelServer interface {
	Send(*Response) error
	Recv() (*CommunicationRequest, error)
	grpc.ServerStream
}

type mafiaInitCommunicationChannelServer struct {
	grpc.ServerStream
}

func (x *mafiaInitCommunicationChannelServer) Send(m *Response) error {
	return x.ServerStream.SendMsg(m)
}

func (x *mafiaInitCommunicationChannelServer) Recv() (*CommunicationRequest, error) {
	m := new(CommunicationRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Mafia_ServiceDesc is the grpc.ServiceDesc for Mafia service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Mafia_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "mafia.Mafia",
	HandlerType: (*MafiaServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "add_user",
			Handler:    _Mafia_AddUser_Handler,
		},
		{
			MethodName: "delete_user",
			Handler:    _Mafia_DeleteUser_Handler,
		},
		{
			MethodName: "get_users",
			Handler:    _Mafia_GetUsers_Handler,
		},
		{
			MethodName: "accuse_user",
			Handler:    _Mafia_AccuseUser_Handler,
		},
		{
			MethodName: "vote_finish_day",
			Handler:    _Mafia_VoteFinishDay_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "init_communication_channel",
			Handler:       _Mafia_InitCommunicationChannel_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "pkg/proto/service/service.proto",
}
