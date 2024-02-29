// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.23.4
// source: booking.proto

package protos

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// BookingServiceClient is the client API for BookingService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BookingServiceClient interface {
	// Public APIs (Note: required user to be logged in to use these APIs)
	Purchase(ctx context.Context, in *PurchaseRequest, opts ...grpc.CallOption) (*Booking, error)
	// Admin APIs
	GetBookingsBySection(ctx context.Context, in *GetBookingsBySectionRequest, opts ...grpc.CallOption) (BookingService_GetBookingsBySectionClient, error)
	RemoveUserFromTrain(ctx context.Context, in *User, opts ...grpc.CallOption) (*emptypb.Empty, error)
	ModifySeat(ctx context.Context, in *ModifySeatRequest, opts ...grpc.CallOption) (*Booking, error)
}

type bookingServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewBookingServiceClient(cc grpc.ClientConnInterface) BookingServiceClient {
	return &bookingServiceClient{cc}
}

func (c *bookingServiceClient) Purchase(ctx context.Context, in *PurchaseRequest, opts ...grpc.CallOption) (*Booking, error) {
	out := new(Booking)
	err := c.cc.Invoke(ctx, "/BookingService/Purchase", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bookingServiceClient) GetBookingsBySection(ctx context.Context, in *GetBookingsBySectionRequest, opts ...grpc.CallOption) (BookingService_GetBookingsBySectionClient, error) {
	stream, err := c.cc.NewStream(ctx, &BookingService_ServiceDesc.Streams[0], "/BookingService/GetBookingsBySection", opts...)
	if err != nil {
		return nil, err
	}
	x := &bookingServiceGetBookingsBySectionClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type BookingService_GetBookingsBySectionClient interface {
	Recv() (*Booking, error)
	grpc.ClientStream
}

type bookingServiceGetBookingsBySectionClient struct {
	grpc.ClientStream
}

func (x *bookingServiceGetBookingsBySectionClient) Recv() (*Booking, error) {
	m := new(Booking)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *bookingServiceClient) RemoveUserFromTrain(ctx context.Context, in *User, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/BookingService/RemoveUserFromTrain", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bookingServiceClient) ModifySeat(ctx context.Context, in *ModifySeatRequest, opts ...grpc.CallOption) (*Booking, error) {
	out := new(Booking)
	err := c.cc.Invoke(ctx, "/BookingService/ModifySeat", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BookingServiceServer is the server API for BookingService service.
// All implementations must embed UnimplementedBookingServiceServer
// for forward compatibility
type BookingServiceServer interface {
	// Public APIs (Note: required user to be logged in to use these APIs)
	Purchase(context.Context, *PurchaseRequest) (*Booking, error)
	// Admin APIs
	GetBookingsBySection(*GetBookingsBySectionRequest, BookingService_GetBookingsBySectionServer) error
	RemoveUserFromTrain(context.Context, *User) (*emptypb.Empty, error)
	ModifySeat(context.Context, *ModifySeatRequest) (*Booking, error)
	mustEmbedUnimplementedBookingServiceServer()
}

// UnimplementedBookingServiceServer must be embedded to have forward compatible implementations.
type UnimplementedBookingServiceServer struct {
}

func (UnimplementedBookingServiceServer) Purchase(context.Context, *PurchaseRequest) (*Booking, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Purchase not implemented")
}
func (UnimplementedBookingServiceServer) GetBookingsBySection(*GetBookingsBySectionRequest, BookingService_GetBookingsBySectionServer) error {
	return status.Errorf(codes.Unimplemented, "method GetBookingsBySection not implemented")
}
func (UnimplementedBookingServiceServer) RemoveUserFromTrain(context.Context, *User) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveUserFromTrain not implemented")
}
func (UnimplementedBookingServiceServer) ModifySeat(context.Context, *ModifySeatRequest) (*Booking, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ModifySeat not implemented")
}
func (UnimplementedBookingServiceServer) mustEmbedUnimplementedBookingServiceServer() {}

// UnsafeBookingServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BookingServiceServer will
// result in compilation errors.
type UnsafeBookingServiceServer interface {
	mustEmbedUnimplementedBookingServiceServer()
}

func RegisterBookingServiceServer(s grpc.ServiceRegistrar, srv BookingServiceServer) {
	s.RegisterService(&BookingService_ServiceDesc, srv)
}

func _BookingService_Purchase_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PurchaseRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BookingServiceServer).Purchase(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/BookingService/Purchase",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BookingServiceServer).Purchase(ctx, req.(*PurchaseRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BookingService_GetBookingsBySection_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetBookingsBySectionRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(BookingServiceServer).GetBookingsBySection(m, &bookingServiceGetBookingsBySectionServer{stream})
}

type BookingService_GetBookingsBySectionServer interface {
	Send(*Booking) error
	grpc.ServerStream
}

type bookingServiceGetBookingsBySectionServer struct {
	grpc.ServerStream
}

func (x *bookingServiceGetBookingsBySectionServer) Send(m *Booking) error {
	return x.ServerStream.SendMsg(m)
}

func _BookingService_RemoveUserFromTrain_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(User)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BookingServiceServer).RemoveUserFromTrain(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/BookingService/RemoveUserFromTrain",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BookingServiceServer).RemoveUserFromTrain(ctx, req.(*User))
	}
	return interceptor(ctx, in, info, handler)
}

func _BookingService_ModifySeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ModifySeatRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BookingServiceServer).ModifySeat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/BookingService/ModifySeat",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BookingServiceServer).ModifySeat(ctx, req.(*ModifySeatRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// BookingService_ServiceDesc is the grpc.ServiceDesc for BookingService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var BookingService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "BookingService",
	HandlerType: (*BookingServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Purchase",
			Handler:    _BookingService_Purchase_Handler,
		},
		{
			MethodName: "RemoveUserFromTrain",
			Handler:    _BookingService_RemoveUserFromTrain_Handler,
		},
		{
			MethodName: "ModifySeat",
			Handler:    _BookingService_ModifySeat_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetBookingsBySection",
			Handler:       _BookingService_GetBookingsBySection_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "booking.proto",
}