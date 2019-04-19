package coordinator

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math/rand"
	"time"
	"mp3/utils"
)

type Coordinator struct {
}

// CoordinatorServer is the coordinator API for Coordinator service.
type CoordinatorServer interface {
	OpenTransaction(context.Context, *Empty) (*Transaction, error)
	CloseTransaction(context.Context, *Transaction) (*Feedback, error)
	CommitTransaction(context.Context, *Empty) (*Feedback, error)
	AbortTransaction(context.Context, *Empty) (*Feedback, error)
}


func (*Coordinator) OpenTransaction(ctx context.Context, req *Empty) (*Transaction, error) {
	transactionID := utils.Concatenate(rand.Intn(1000000), int(time.Now().Unix()))
	return &Transaction{Id:transactionID}, nil
}

func (*Coordinator) CloseTransaction(ctx context.Context, req *Transaction) (*Feedback, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CloseTransaction not implemented")
}

func (*Coordinator) CommitTransaction(ctx context.Context, req *Empty) (*Feedback, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CommitTransaction not implemented")
}

func (*Coordinator) AbortTransaction(ctx context.Context, req *Empty) (*Feedback, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AbortTransaction not implemented")
}