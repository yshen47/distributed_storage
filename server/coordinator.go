package server

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math/rand"
	"mp3/utils"
	"time"
)

// Coordinator can be embedded to have forward compatible implementations.
type Coordinator struct {

}

func (*Coordinator) OpenTransaction(ctx context.Context, req *Empty) (*Transaction, error) {
	transactionID := utils.Concatenate(rand.Intn(1000000), int(time.Now().Unix()))
	return &Transaction{Id:&transactionID}, nil

}
func (*Coordinator) CloseTransaction(ctx context.Context, req *Transaction) (*Feedback, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CloseTransaction not implemented")
}

func (*Coordinator) AskCommitTransaction(ctx context.Context, req *Transaction) (*Feedback, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AskCommitTransaction not implemented")
}
func (*Coordinator) AskAbortTransaction(ctx context.Context, req *Transaction) (*Feedback, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AskAbortTransaction not implemented")
}
func (*Coordinator) TryLock(ctx context.Context, req *TryLockParam) (*Feedback, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TryLock not implemented")
}
func (*Coordinator) ReportUnlock(ctx context.Context, req *ReportUnLockParam) (*Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReportUnlock not implemented")
}
