package wsim

import (
	"context"

	"github.com/caiwp/wsim/api/pb"
)

// IRpcClient 请求RPC服务
// 扩展实现
type IRpcClient interface {
	// 鉴权
	Auth(ctx context.Context, req *pb.AuthReq) (*pb.AuthReply, error)
	// 操作
	Operate(ctx context.Context, req *pb.Proto) error
}

// IRpcServer RPC服务
// 扩展实现
type IRpcServer interface {
	// 推送单条
	PushOne(ctx context.Context, req *pb.PushOneReq, reply *pb.PushOneReply) error
	// 批量推送
	PushBatch(ctx context.Context, req *pb.PushBatchReq, reply *pb.PushBatchReply) error
	// 广播房间
	BroadcaseRoom(ctx context.Context, req *pb.BroadcastRoomReq, reply *pb.BroadcastRoomReply) error
}
