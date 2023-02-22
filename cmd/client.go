package main

import (
	"context"

	"github.com/caiwp/wsim"
	"github.com/caiwp/wsim/api/pb"
	"go.uber.org/zap"
)

var _ wsim.IRpcClient = (*RpcClient)(nil)

type RpcClient struct {
	logger *zap.Logger
}

func NewRpcClient(logger *zap.Logger) *RpcClient {
	return &RpcClient{
		logger: logger,
	}
}

// 鉴权
func (c *RpcClient) Auth(ctx context.Context, req *pb.AuthReq) (*pb.AuthReply, error) {
	c.logger.Info("Auth", zap.Any("req", req))
	// TODO: auth token
	return &pb.AuthReply{
		Cid: req.Token,
	}, nil
}

// 操作
func (c *RpcClient) Operate(ctx context.Context, req *pb.Proto) error {
	c.logger.Info("Operate", zap.Any("req", req))
	return nil
}
