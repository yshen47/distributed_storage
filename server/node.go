package server

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"context"
)

// NodeServer is the server API for Node service.
type NodeServer interface {
	ClientSet(context.Context, *SetParams) (*Feedback, error)
	ClientGet(context.Context, *GetParams) (*Transaction, error)
}

// Node can be embedded to have forward compatible implementations.
type Node struct {
}

func (*Node) ClientSet(ctx context.Context, req *SetParams) (*Feedback, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ClientSet not implemented")
}
func (*Node) ClientGet(ctx context.Context, req *GetParams) (*Transaction, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ClientGet not implemented")
}