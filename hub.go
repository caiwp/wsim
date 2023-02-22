package wsim

import (
	"log"
	"sync"
)

type Hub struct {
	hid        string
	clientMap  sync.Map
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	clearDelay int
	done       chan struct{}
}

func NewHub(hid string) *Hub {
	h := &Hub{
		hid:        hid,
		clientMap:  sync.Map{},
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		done:       make(chan struct{}, 1),
	}

	go h.run()
	return h
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clientMap.Store(client, true)

		case client := <-h.unregister:
			if _, ok := h.clientMap.Load(client); ok {
				h.clientMap.Delete(client)
				close(client.send)
			}

		case message := <-h.broadcast:
			h.clientMap.Range(func(key, value interface{}) bool {
				client, ok := key.(*Client)
				if ok {
					select {
					case client.send <- message:
					default: // 下线
						close(client.send)
						h.clientMap.Delete(client)
					}
				}

				return true
			})
		case <-h.done:
			log.Println("done: ", h.hid)
			return
		}
	}
}

func (h *Hub) ClientNum() (num int) {
	h.clientMap.Range(func(key, value interface{}) bool {
		num++
		return true
	})
	return
}

func (h *Hub) Close() {
	close(h.done)
}
