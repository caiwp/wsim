package wsim

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

var bucket sync.Map // rid:*room

func GetOrCreateRoom(rid string) *Room {
	v, ok := bucket.Load(rid)
	if ok {
		return v.(*Room)
	}

	room := NewRoom(rid)
	bucket.Store(rid, room)
	return room
}

func GetRoom(rid string) *Room {
	v, ok := bucket.Load(rid)
	if ok {
		return v.(*Room)
	}
	return nil
}

// RunTick 定时维护 bucket
func RunTick(done <-chan struct{}, logger *zap.Logger) {
	var tk = time.NewTicker(1 * time.Minute) // 每分钟
	defer tk.Stop()

	for {
		select {
		case <-tk.C:
			logger.Info("tick...")
			bucket.Range(func(key, value interface{}) bool {
				room, ok := value.(*Room)
				if !ok {
					return true
				}

				// 清除空闲房间
				if room.clearDelay >= clearDelyNum {
					logger.Info("clear", zap.String("rid", room.rid))
					room.Close()
					bucket.Delete(key)
					return true
				}

				num := room.ClientNum()
				if num == 0 {
					room.clearDelay++
				} else {
					room.clearDelay = 0 // 重置
				}

				logger.Info("ClientNum", zap.Int("num", num), zap.String("rid", room.rid))
				return true
			})

		case <-done:
			logger.Info("RunTick done")
			return
		}
	}
}
