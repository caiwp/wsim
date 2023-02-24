package wsim

import (
	"fmt"
	"net"

	"github.com/caiwp/wsim/api/pb"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func RunRpcServer(port int, logger *zap.Logger) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	logger.Sugar().Infof("rpc server listening at %v", lis.Addr())

	srv := grpc.NewServer()
	pb.RegisterWsImServer(srv, NewServer(logger))
	return srv.Serve(lis)
}
