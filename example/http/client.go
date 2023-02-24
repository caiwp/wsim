package main

import (
	"context"
	"time"

	"github.com/caiwp/wsim"
	"github.com/caiwp/wsim/api/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var _ wsim.IRpcClient = (*RpcClient)(nil)

// 将逻辑引回自身业务处理

type RpcClient struct {
	rid    uint32
	logger *zap.Logger
	cl     pb.WsImClient
	conn   *grpc.ClientConn
}

func NewRpcClient(serverAddr string, logger *zap.Logger) (*RpcClient, error) {
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &RpcClient{
		logger: logger,
		cl:     pb.NewWsImClient(conn),
		conn:   conn,
	}, nil
}

// 鉴权
func (c *RpcClient) Auth(ctx context.Context, req *pb.AuthReq) (*pb.AuthReply, error) {
	c.logger.Info("Auth", zap.Any("req", req))

	c.rid = req.Rid
	// TODO: auth token
	return &pb.AuthReply{
		Cid: uint32(time.Now().Unix()),
	}, nil
}

// 操作
func (c *RpcClient) Operate(ctx context.Context, req *pb.Proto) error {
	c.logger.Info("Operate", zap.Any("req", req))

	// TODO: 根据自己需要调整
	_, err := c.cl.BroadcastRoom(ctx, &pb.BroadcastRoomReq{
		Rid: c.rid,
		Proto: &pb.Proto{
			Ver:  req.Ver,
			Op:   req.Op,
			Seq:  req.Seq,
			Body: req.Body,
		},
	})
	return err
}

func (c *RpcClient) Close() {
	c.conn.Close()
}
