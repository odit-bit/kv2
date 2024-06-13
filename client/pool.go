package client

import (
	"errors"
	"net"
	"sync"
)

var ErrMaxConnection = errors.New("reach max connection")

type connPool struct {
	connC chan net.Conn
	New   func() (net.Conn, error)
	Min   int
	Max   int
	mx    sync.Mutex
}

func NewConnPool(min, max int, fn func() (net.Conn, error)) (*connPool, error) {
	if min < 0 || max <= 0 || min > max {
		return nil, errors.New("invalid min or max settings")
	}

	pool := &connPool{
		connC: make(chan net.Conn, min),
		New:   fn,
		Min:   min,
		Max:   max,
		mx:    sync.Mutex{},
	}

	conn, err := pool.New()
	if err != nil {
		return nil, err
	}
	pool.connC <- conn
	return pool, nil

}

func (cp *connPool) Get() (net.Conn, error) {
	select {
	case conn := <-cp.connC:
		return conn, nil
	default:
		cp.mx.Lock()
		defer cp.mx.Unlock()

		if len(cp.connC) < cp.Max {
			return cp.New()
		}

		return nil, ErrMaxConnection
	}

}

func (cp *connPool) Put(conn net.Conn) {
	select {
	case cp.connC <- conn:
	default:
		conn.Close()
	}
}

func (cp *connPool) Close() {
	cp.mx.Lock()
	defer cp.mx.Unlock()

	close(cp.connC)
	for conn := range cp.connC {
		conn.Close()
	}
}

// func (cp *connPool) Get() net.Conn {
// 	conn := <-cp.connC
// 	return conn
// }

// func (cp *connPool) Put(conn net.Conn) {
// 	cp.connC <- conn
// }
