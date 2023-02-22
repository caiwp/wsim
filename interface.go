package wsim

import (
	"context"

	"github.com/caiwp/wsim/api/pb"
)

type IRpcClient interface {
	// 鉴权
	Auth(ctx context.Context, req *pb.AuthReq) (*pb.AuthReply, error)
	// 操作
	Operate(ctx context.Context, req *pb.Proto) error
}
