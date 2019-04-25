package server

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"context"
)

// Node can be embedded to have forward compatible implementations.
type Node struct {
	name string
	data map[string]string
}

// NodeServer is the server API for Node service.
type NodeServer interface {
	ClientSet(context.Context, *SetParams) (*Feedback, error)
	ClientGet(context.Context, *GetParams) (*Transaction, error)
}

func (n *Node) ClientSet(ctx context.Context, req *SetParams) (*Feedback, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ClientSet not implemented")
}
func (n *Node) ClientGet(ctx context.Context, req *GetParams) (*Feedback, error) {
	if *req.ServerIdentifier != n.name {
		return nil, status.Error(codes.InvalidArgument, "call the wrong node server")
	}

	//TODO: add lock

	resFeedback := &Feedback{}
	var result string

	val, ok := n.data[*req.ObjectName]
	if ok {
		result = val
		resFeedback.Message = &result
		return resFeedback, nil
	} else {
		result = "NOT FOUND"
		resFeedback.Message = &result
		//TODO: tell coordinator to abort the current transaction
		return resFeedback, status.Error(codes.Aborted, "not found")
	}

}