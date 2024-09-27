// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package legacy

import (
	context "context"
	v1 "github.com/llm-operator/vector-store-manager/api/v1"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// VectorStoreInternalServiceClient is the client API for VectorStoreInternalService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type VectorStoreInternalServiceClient interface {
	SearchVectorStore(ctx context.Context, in *v1.SearchVectorStoreRequest, opts ...grpc.CallOption) (*v1.SearchVectorStoreResponse, error)
}

type vectorStoreInternalServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewVectorStoreInternalServiceClient(cc grpc.ClientConnInterface) VectorStoreInternalServiceClient {
	return &vectorStoreInternalServiceClient{cc}
}

func (c *vectorStoreInternalServiceClient) SearchVectorStore(ctx context.Context, in *v1.SearchVectorStoreRequest, opts ...grpc.CallOption) (*v1.SearchVectorStoreResponse, error) {
	out := new(v1.SearchVectorStoreResponse)
	err := c.cc.Invoke(ctx, "/llmoperator.vector_store.v1.VectorStoreInternalService/SearchVectorStore", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// VectorStoreInternalServiceServer is the server API for VectorStoreInternalService service.
// All implementations must embed UnimplementedVectorStoreInternalServiceServer
// for forward compatibility
type VectorStoreInternalServiceServer interface {
	SearchVectorStore(context.Context, *v1.SearchVectorStoreRequest) (*v1.SearchVectorStoreResponse, error)
	mustEmbedUnimplementedVectorStoreInternalServiceServer()
}

// UnimplementedVectorStoreInternalServiceServer must be embedded to have forward compatible implementations.
type UnimplementedVectorStoreInternalServiceServer struct {
}

func (UnimplementedVectorStoreInternalServiceServer) SearchVectorStore(context.Context, *v1.SearchVectorStoreRequest) (*v1.SearchVectorStoreResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SearchVectorStore not implemented")
}
func (UnimplementedVectorStoreInternalServiceServer) mustEmbedUnimplementedVectorStoreInternalServiceServer() {
}

// UnsafeVectorStoreInternalServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to VectorStoreInternalServiceServer will
// result in compilation errors.
type UnsafeVectorStoreInternalServiceServer interface {
	mustEmbedUnimplementedVectorStoreInternalServiceServer()
}

func RegisterVectorStoreInternalServiceServer(s grpc.ServiceRegistrar, srv VectorStoreInternalServiceServer) {
	s.RegisterService(&VectorStoreInternalService_ServiceDesc, srv)
}

func _VectorStoreInternalService_SearchVectorStore_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1.SearchVectorStoreRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(VectorStoreInternalServiceServer).SearchVectorStore(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/llmoperator.vector_store.v1.VectorStoreInternalService/SearchVectorStore",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(VectorStoreInternalServiceServer).SearchVectorStore(ctx, req.(*v1.SearchVectorStoreRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// VectorStoreInternalService_ServiceDesc is the grpc.ServiceDesc for VectorStoreInternalService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var VectorStoreInternalService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "llmoperator.vector_store.v1.VectorStoreInternalService",
	HandlerType: (*VectorStoreInternalServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SearchVectorStore",
			Handler:    _VectorStoreInternalService_SearchVectorStore_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/v1/legacy/vector_store.proto",
}