package wsim

import (
	"context"

	"github.com/caiwp/wsim/api/pb"
	"go.uber.org/zap"
)

type Server struct {
	logger *zap.Logger
	pb.UnimplementedWsImServer
}

func NewServer(logger *zap.Logger) *Server {
	return &Server{
		logger: logger,
	}
}

// 推送单条
func (s *Server) PushOne(ctx context.Context, req *pb.PushOneReq) (reply *pb.PushOneReply, err error) {
	s.logger.Info("PushOne", zap.Any("req", req))
	reply = &pb.PushOneReply{}

	room := GetRoom(req.Rid)
	if room == nil {
		return nil, ErrRoomNotFound
	}

	client := room.GetClient(req.Cid)
	if client == nil {
		return nil, ErrClientNotFound
	}

	if !client.Push(req.Proto) {
		return nil, ErrMessageFull
	}

	return
}

// 批量推送
func (s *Server) PushBatch(ctx context.Context, req *pb.PushBatchReq) (reply *pb.PushBatchReply, err error) {
	s.logger.Info("PushBatch", zap.Any("req", req))
	reply = &pb.PushBatchReply{}

	var tmp = make(map[uint32]map[uint32]*pb.PushBatchReq_Item)
	for _, v := range req.List {
		if len(tmp[v.Rid]) == 0 {
			tmp[v.Rid] = make(map[uint32]*pb.PushBatchReq_Item)
		}
		tmp[v.Rid][v.Cid] = v
	}

	for rid, v := range tmp {
		room := GetRoom(rid)
		if room == nil { // 房间不存在
			continue
		}

		for cid, vv := range v {
			client := room.GetClient(cid)
			if client == nil {
				continue
			}

			client.Push(vv.Proto)
		}
	}

	return
}

// 广播房间
func (s *Server) BroadcastRoom(ctx context.Context, req *pb.BroadcastRoomReq) (reply *pb.BroadcastRoomReply, err error) {
	s.logger.Info("BroadcastRoom", zap.Any("req", req))
	reply = &pb.BroadcastRoomReply{}

	room := GetRoom(req.Rid)
	if room == nil {
		return nil, ErrRoomNotFound
	}

	room.BroadcastRoom(req.Proto)
	return
}

// 服务信息
func (s *Server) ServerInfo(ctx context.Context, req *pb.ServerInfoReq) (reply *pb.ServerInfoReply, err error) {
	s.logger.Info("ServerInfo", zap.Any("req", req))
	reply = &pb.ServerInfoReply{}

	if len(req.Rids) > 0 {
		for _, v := range req.Rids {
			room := GetRoom(v)
			if room != nil {
				reply.List = append(reply.List, &pb.ServerInfoReply_Item{
					Rid: v,
					Num: uint32(room.num),
				})
			}
		}
		return
	}

	// 全部
	bucket.Range(func(key, value interface{}) bool {
		room, ok := value.(*Room)
		if ok {
			reply.List = append(reply.List, &pb.ServerInfoReply_Item{
				Rid: room.rid,
				Num: uint32(room.num),
			})
		}
		return true
	})

	return
}
