package server

import (
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"context"
)

// Node can be embedded to have forward compatible implementations.
type Node struct {
	Name                string
	CoordinatorDelegate CoordinatorClient
	data                map[string]string
}

// NodeServer is the server API for Node service.
type NodeServer interface {
	ClientSet(context.Context, *SetParams) (*Feedback, error)
	ClientGet(context.Context, *GetParams) (*Feedback, error)
}

func (n *Node) ClientSet(ctx context.Context, req *SetParams) (*Feedback, error) {
	if *req.ServerIdentifier != n.Name {
		name := fmt.Sprintf("request is %s, myname is %s",*req.ServerIdentifier,n.Name)
		return nil, status.Error(codes.InvalidArgument, name)
	}
	n.data[*req.ObjectName] = *req.Value
	resFeedback := &Feedback{}
	res := "OK"
	resFeedback.Message = &res
	return resFeedback, nil
}

func (n *Node) ClientGet(ctx context.Context, req *GetParams) (*Feedback, error) {
	if *req.ServerIdentifier != n.Name {
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