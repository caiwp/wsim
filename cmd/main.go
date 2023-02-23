package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/pprof"

	"github.com/caiwp/wsim"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	addr = flag.String("addr", ":8080", "http service address")
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

		hub := wsim.GetOrCreateRoom(roomId)
		wsim.ServerWs(hub, ctx.Writer, ctx.Request, NewRpcClient(logger), logger)
	})

	if err := app.Run(*addr); err != nil {
		panic(err)
	}
}
