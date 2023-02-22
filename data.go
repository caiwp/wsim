package wsim

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

var hubMap sync.Map

func GetOrCreateHub(hid string) *Hub {
	v, ok := hubMap.Load(hid)
	if ok {
		return v.(*Hub)
	}

	hub := NewHub(hid)
	hubMap.Store(hid, hub)
	return hub
}

func GetHub(hid string) *Hub {
	v, ok := hubMap.Load(hid)
	if ok {
		return v.(*Hub)
	}
	return nil
}

// RunTick 定时维护 hubMap
func RunTick(done <-chan struct{}, logger *zap.Logger) {
	var tk = time.NewTicker(1 * time.Minute) // 每分钟
	defer tk.Stop()

	for {
		select {
		case <-tk.C:
			logger.Info("tick...")
			hubMap.Range(func(key, value interface{}) bool {
				hub, ok := value.(*Hub)
				if !ok {
					return true
				}

				if hub.clearDelay >= clearDelyNum {
					hub.Close()
					hubMap.Delete(key)
					return true
				}

				if hub.ClientNum() == 0 {
					hub.clearDelay++
				} else {
					hub.clearDelay = 0 // 重置
				}

				return true
			})

		case <-done:
			logger.Info("RunTick done")
			return
		}
	}
}
