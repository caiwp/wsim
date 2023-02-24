package wsim

import (
	"sync"

	"github.com/caiwp/wsim/api/pb"
	"go.uber.org/zap"
)

type Room struct {
	rid        uint32
	clientMap  sync.Map // cid *Client
	broadcast  chan *pb.Proto
	register   chan *Client
	unregister chan *Client
	clearDelay int
	done       chan struct{}
	num        int
	logger     *zap.Logger
}

func NewRoom(rid uint32, logger *zap.Logger) *Room {
	r := &Room{
		rid:        rid,
		clientMap:  sync.Map{},
		broadcast:  make(chan *pb.Proto, 256),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
		done:       make(chan struct{}, 1),
		logger:     logger.With(zap.Uint32("rid", rid)),
	}

	go r.run()
	return r
}

func (r *Room) run() {
	for {
		select {
		case client := <-r.register:
			r.logger.Debug("register", zap.Uint32("client", client.cid))
			r.clientMap.Store(client.cid, client)

		case client := <-r.unregister:
			r.logger.Debug("unregister", zap.Uint32("client", client.cid))
			if _, ok := r.clientMap.Load(client.cid); ok {
				r.clientMap.Delete(client.cid)
				client.Close()
			}

		case msg := <-r.broadcast:
			r.clientMap.Range(func(key, value interface{}) bool {
				client, ok := value.(*Client)
				if ok {
					if !client.Push(msg) { // 上线了
						client.Close()
						r.clientMap.Delete(client.cid)
					}
				}

				return true
			})

		case <-r.done:
			r.logger.Info("room done")
			return
		}
	}
}

func (r *Room) BroadcastRoom(msg *pb.Proto) {
	r.broadcast <- msg
}

func (r *Room) clientNum() (num int) {
	r.clientMap.Range(func(key, value interface{}) bool {
		num++
		return true
	})
	r.num = num
	return
}

func (r *Room) Close() {
	close(r.done)
}

func (r *Room) GetClient(cid uint32) *Client {
	v, ok := r.clientMap.Load(cid)
	if ok {
		return v.(*Client)
	}
	return nil
}
