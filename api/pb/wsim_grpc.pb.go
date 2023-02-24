// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.15.5
// source: proto/wsim.proto

package pb

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

// WsImClient is the client API for WsIm service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type WsImClient interface {
	// 推送个人
	// rpc PushOne(PushOneReq) returns (PushOneReply) {}
	// 批量推送
	// rpc PushBatch(PushBatchReq) returns (PushBatchReply) {}
	// 广播房间
	BroadcastRoom(ctx context.Context, in *BroadcastRoomReq, opts ...grpc.CallOption) (*BroadcastRoomReply, error)
}

type wsImClient struct {
	cc grpc.ClientConnInterface
}

func NewWsImClient(cc grpc.ClientConnInterface) WsImClient {
	return &wsImClient{cc}
}

func (c *wsImClient) BroadcastRoom(ctx context.Context, in *BroadcastRoomReq, opts ...grpc.CallOption) (*BroadcastRoomReply, error) {
	out := new(BroadcastRoomReply)
	err := c.cc.Invoke(ctx, "/pb.WsIm/BroadcastRoom", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WsImServer is the server API for WsIm service.
// All implementations must embed UnimplementedWsImServer
// for forward compatibility
type WsImServer interface {
	// 推送个人
	// rpc PushOne(PushOneReq) returns (PushOneReply) {}
	// 批量推送
	// rpc PushBatch(PushBatchReq) returns (PushBatchReply) {}
	// 广播房间
	BroadcastRoom(context.Context, *BroadcastRoomReq) (*BroadcastRoomReply, error)
	mustEmbedUnimplementedWsImServer()
}

// UnimplementedWsImServer must be embedded to have forward compatible implementations.
type UnimplementedWsImServer struct {
}

func (UnimplementedWsImServer) BroadcastRoom(context.Context, *BroadcastRoomReq) (*BroadcastRoomReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BroadcastRoom not implemented")
}
func (UnimplementedWsImServer) mustEmbedUnimplementedWsImServer() {}

// UnsafeWsImServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to WsImServer will
// result in compilation errors.
type UnsafeWsImServer interface {
	mustEmbedUnimplementedWsImServer()
}

func RegisterWsImServer(s grpc.ServiceRegistrar, srv WsImServer) {
	s.RegisterService(&WsIm_ServiceDesc, srv)
}

func _WsIm_BroadcastRoom_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BroadcastRoomReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WsImServer).BroadcastRoom(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.WsIm/BroadcastRoom",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WsImServer).BroadcastRoom(ctx, req.(*BroadcastRoomReq))
	}
	return interceptor(ctx, in, info, handler)
}

// WsIm_ServiceDesc is the grpc.ServiceDesc for WsIm service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var WsIm_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pb.WsIm",
	HandlerType: (*WsImServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "BroadcastRoom",
			Handler:    _WsIm_BroadcastRoom_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/wsim.proto",
}