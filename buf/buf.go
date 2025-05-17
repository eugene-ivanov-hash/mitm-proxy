package buf

import (
	"sync"
)

var (
	bytePool = sync.Pool{
		New: func() interface{} {
			return []byte{}
		},
	}
	byteSliceChan = make(chan []byte, 10)
)

func ByteGet(length int) (data []byte) {
	select {
	case data = <-byteSliceChan:
	default:
		data = bytePool.Get().([]byte)[:0]
	}

	if cap(data) < length {
		data = make([]byte, length)
	} else {
		data = data[:length]
	}

	return data
}

func BytePut(data []byte) {
	select {
	case byteSliceChan <- data:
	default:
		bytePool.Put(data)
	}
}
