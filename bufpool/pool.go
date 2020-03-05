package bufpool

import (
	"bytes"
	"sync"
)

var (
	pool = sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(nil)
		},
	}
)

func Get() *bytes.Buffer {
	return pool.Get().(*bytes.Buffer)
}

func Release(buffer *bytes.Buffer) {
	buffer.Reset()
	pool.Put(buffer)
}
