package wsim

import (
	"log"
	"sync"

	"github.com/caiwp/wsim/api/pb"
)

type Room struct {
	rid        string
	clientMap  sync.Map
	broadcast  chan *pb.Proto
	register   chan *Client
	unregister chan *Client
	clearDelay int
	done       chan struct{}
}

func NewRoom(rid string) *Room {
	r := &Room{
		rid:        rid,
		clientMap:  sync.Map{},
		broadcast:  make(chan *pb.Proto),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		done:       make(chan struct{}, 1),
	}

	go r.run()
	return r
}

func (r *Room) run() {
	for {
		select {
		case client := <-r.register:
			r.clientMap.Store(client, true)

		case client := <-r.unregister:
			if _, ok := r.clientMap.Load(client); ok {
				r.clientMap.Delete(client)
				close(client.send)
			}

		case message := <-r.broadcast:
			r.clientMap.Range(func(key, value interface{}) bool {
				client, ok := key.(*Client)
				if ok {
					select {
					case client.send <- message:
					default: // 下线
						close(client.send)
						r.clientMap.Delete(client)
					}
				}

				return true
			})
		case <-r.done:
			log.Println("done: ", r.rid)
			return
		}
	}
}

func (r *Room) broadcastRoom(proto *pb.Proto) {
	r.broadcast <- proto
}

func (r *Room) ClientNum() (num int) {
	r.clientMap.Range(func(key, value interface{}) bool {
		num++
		return true
	})
	return
}

func (r *Room) Close() {
	close(r.done)
}
