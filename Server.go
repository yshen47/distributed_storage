package clientServer

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"context"
)

type UnimplementedServerServer struct {

}

func (*UnimplementedServerServer) ClientSet(ctx context.Context, req *SetParams) (*Status, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ClientSet not implemented")
}
func (*UnimplementedServerServer) ClientGet(ctx context.Context, req *GetParams) (*Transaction, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ClientGet not implemented")
}