package wsim

import (
	"bytes"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	logger *zap.Logger
}

func NewClient(hub *Hub, conn *websocket.Conn, logger *zap.Logger) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 256),
		logger: logger.With(zap.String("conn", conn.RemoteAddr().String())),
	}
}

func (c *Client) readPump() {
	defer func() {
		c.close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
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
		// FIXME: msg send
		c.hub.broadcast <- msg
	}
}

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
			w.Write(msg)

			// 批量写入
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
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

func (c *Client) close() {
	c.hub.unregister <- c
	c.conn.Close()
}

func ServerWs(hub *Hub, w http.ResponseWriter, r *http.Request, logger *zap.Logger) {
	logger.Debug("ServerWs")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("ServerWs", zap.Error(err))
		return
	}

	logger.Info("ServerWs", zap.String("conn", conn.RemoteAddr().String()))
	client := NewClient(hub, conn, logger)
	client.hub.register <- client

	go client.readPump()
	go client.writePump()
}
