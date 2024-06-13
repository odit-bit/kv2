package store

import (
	"bytes"
	"sync"
)

var bpool = newBBufpool()

type bbufPool struct {
	bufPool sync.Pool
}

func newBBufpool() *bbufPool {
	pool := bbufPool{
		bufPool: sync.Pool{
			New: func() any {
				buff := &bytes.Buffer{}
				return buff
			},
		},
	}

	return &pool
}

func (pool *bbufPool) Get() *bytes.Buffer {
	v := pool.bufPool.Get()
	return v.(*bytes.Buffer)
}

func (pool *bbufPool) Put(buf *bytes.Buffer) {
	buf.Reset()
	pool.bufPool.Put(buf)
}
