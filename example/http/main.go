package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-contrib/pprof"

	"github.com/caiwp/wsim"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	addr    = flag.String("addr", ":8080", "http server address")
	port    = flag.Int("port", 50051, "rpc server port")
	rpcAddr = flag.String("rpcAddr", "localhost:50051", "rpc server address")
)

func main() {
	flag.Parse()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	done := make(chan struct{}, 1)
	go wsim.RunTick(done, logger)

	go func() {
		signs := make(chan os.Signal, 1)
		signal.Notify(signs, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(signs)

		sign := <-signs
		logger.Info("receive sign", zap.Int("pid", os.Getpid()), zap.Any("sign", sign))

		close(done)
	}()

	go runHttp(logger)
	go runRpc(logger)

	<-done

	time.Sleep(1 * time.Second)
}

func runHttp(logger *zap.Logger) {
	app := gin.New()
	pprof.Register(app)

	app.LoadHTMLFiles("index.html")

	app.GET("/room/:roomId", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "index.html", nil)
	})

	app.GET("/ws/:roomId", func(ctx *gin.Context) {
		roomId := ctx.Param("roomId")
		logger.Debug("ws", zap.String("roomId", roomId))

		rid, err := strconv.ParseInt(roomId, 10, 64)
		if err != nil {
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}

		cl, err := NewRpcClient(*rpcAddr, logger)
		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		hub := wsim.GetOrCreateRoom(uint32(rid), logger)
		wsim.ServerWs(hub, ctx.Writer, ctx.Request, cl, logger)
	})

	if err := app.Run(*addr); err != nil {
		panic(err)
	}
}

func runRpc(logger *zap.Logger) {
	err := wsim.RunRpcServer(*port, logger)
	if err != nil {
		panic(err)
	}
}
