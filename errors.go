package wsim

import "errors"

var (
	ErrRoomNotFound   = errors.New("wsim: room not found")
	ErrClientNotFound = errors.New("wsim: client not found")
	ErrMessageFull    = errors.New("wsim: message full")
)
