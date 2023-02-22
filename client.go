package wsim

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/caiwp/wsim/api/pb"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Client struct {
	cid       string
	room      *Room
	conn      *websocket.Conn
	send      chan *pb.Proto
	logger    *zap.Logger
	rpcClient IRpcClient
}

func NewClient(room *Room, conn *websocket.Conn, rpcClient IRpcClient, logger *zap.Logger) *Client {
	return &Client{
		room:      room,
		conn:      conn,
		send:      make(chan *pb.Proto, 256),
		logger:    logger.With(zap.String("conn", conn.RemoteAddr().String())),
		rpcClient: rpcClient,
	}
}

// 接收客户端数据
func (c *Client) readPump() {
	defer func() {
		c.room.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.logger.Debug("pong")
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error("readPump ReadMessage", zap.Error(err))
			}
			break
		}

		msg = bytes.TrimSpace(bytes.Replace(msg, newline, space, -1))
		c.logger.Debug("readPump", zap.ByteString("msg", msg))

		var proto = new(pb.Proto)
		if err = json.Unmarshal(msg, &proto); err != nil {
			c.logger.Warn(string(msg), zap.Error(err))
			continue
		}

		switch proto.Op {
		case int32(pb.Op_AUTH):
			reply, err := c.rpcClient.Auth(context.TODO(), &pb.AuthReq{
				Rid:   c.room.rid,
				Token: proto.Body,
			})
			if err != nil || len(reply.Cid) == 0 { // 鉴权失败
				c.logger.Error("Auth failed", zap.Error(err), zap.Any("reply", reply))
				break
			}
			c.setCid(reply.Cid)

		case int32(pb.Op_SEND): // 必须先鉴权
			if !c.hadAuthed() {
				c.logger.Warn("no auth", zap.Any("proto", proto))
				continue
			}

			if err := c.rpcClient.Operate(context.TODO(), proto); err != nil {
				c.logger.Error("Operate failed", zap.Error(err), zap.Any("proto", proto))
				continue
			}

			// FIXME:
			c.room.broadcastRoom(proto)
		}
	}
}

func (c *Client) setCid(cid string) {
	c.cid = cid
	c.logger = c.logger.With(zap.String("cid", cid))
}

func (c *Client) hadAuthed() bool {
	return len(c.cid) > 0
}

// 推送客户端数据
func (c *Client) writePump() {
	tk := time.NewTicker(pingPeriod)
	defer func() {
		tk.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok { // channel close
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(c.parseMsg(msg))

			// 批量写入
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(c.parseMsg(<-c.send))
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-tk.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			c.logger.Debug("ping")
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) parseMsg(msg *pb.Proto) []byte {
	byt, _ := json.Marshal(msg)
	return byt
}

func ServerWs(room *Room, w http.ResponseWriter, r *http.Request, rpcClient IRpcClient, logger *zap.Logger) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("ServerWs", zap.Error(err))
		return
	}

	logger.Info("ServerWs", zap.String("conn", conn.RemoteAddr().String()))
	client := NewClient(room, conn, rpcClient, logger)
	client.room.register <- client

	go client.readPump()
	go client.writePump()
}
