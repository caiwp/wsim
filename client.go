package wsim

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/caiwp/wsim/api/pb"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Client struct {
	cid       uint32
	room      *Room
	conn      *websocket.Conn
	send      chan *pb.Proto
	logger    *zap.Logger
	rpcClient IRpcClient
	once      *sync.Once
}

func NewClient(room *Room, conn *websocket.Conn, rpcClient IRpcClient, logger *zap.Logger) *Client {
	return &Client{
		room:      room,
		conn:      conn,
		send:      make(chan *pb.Proto, 256),
		logger:    logger.With(zap.String("conn", conn.RemoteAddr().String())),
		rpcClient: rpcClient,
		once:      new(sync.Once),
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
			c.logger.Warn("read", zap.Error(err))
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
		case int32(pb.Op_AUTH): // 鉴权
			reply, err := c.rpcClient.Auth(context.TODO(), &pb.AuthReq{
				Rid:   c.room.rid,
				Token: proto.Body,
			})
			if err != nil || reply.Cid == 0 { // 鉴权失败
				c.logger.Error("Auth failed", zap.Error(err), zap.Any("reply", reply))
				break
			}
			c.init(reply.Cid)

		case int32(pb.Op_SEND): // 必须先鉴权
			if !c.authed() {
				c.logger.Warn("no auth", zap.Any("proto", proto))
				continue
			}

			if err := c.rpcClient.Operate(context.TODO(), proto); err != nil {
				c.logger.Error("Operate failed", zap.Error(err), zap.Any("proto", proto))
				continue
			}
		}
	}
}

func (c *Client) init(cid uint32) {
	c.cid = cid
	c.logger = c.logger.With(zap.Uint32("cid", cid))
	c.room.register <- c
}

func (c *Client) authed() bool {
	return c.cid > 0
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

func (c *Client) Close() {
	c.logger.Debug("close")
	c.once.Do(func() {
		c.logger.Debug("close do")
		c.rpcClient.Close()
		close(c.send)
	})
}

func (c *Client) Push(proto *pb.Proto) (ok bool) {
	select {
	case c.send <- proto:
		return true
	default:
	}
	return
}

func ServerWs(room *Room, w http.ResponseWriter, r *http.Request, rpcClient IRpcClient, logger *zap.Logger) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("ServerWs", zap.Error(err))
		return
	}

	logger.Info("ServerWs", zap.String("conn", conn.RemoteAddr().String()))
	client := NewClient(room, conn, rpcClient, logger)

	go client.readPump()
	go client.writePump()
}
