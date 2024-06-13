package server

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"sync/atomic"
	"time"

	"github.com/odit-bit/kv2/pkg/protocol"
)

type Cacher interface {
	Get(key []byte) ([]byte, bool)
	Set(key, value []byte, ttl int64) error
	Delete(key []byte) error
}

type CacheServer struct {
	addr      string
	l         *net.TCPListener
	cache     Cacher
	currCount atomic.Int64
	logger    *slog.Logger
	// connC     chan net.Conn
}

func New(addr string, db Cacher, logger *slog.Logger) *CacheServer {

	resolve, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	listener, err := net.ListenTCP("tcp", resolve)
	if err != nil {
		log.Fatal(err)
	}

	if logger == nil {
		logger = &slog.Logger{}
	}

	cs := CacheServer{
		addr:      addr,
		l:         listener,
		cache:     db,
		currCount: atomic.Int64{},
		logger:    logger,
	}
	return &cs
}

func (srv *CacheServer) Address() string {
	return srv.l.Addr().String()
}

func (srv *CacheServer) ShutDown() error {
	srv.logger.Info("shutting down server")
	if err := srv.l.Close(); err != nil {
		return err
	}
	return nil
}

func (srv *CacheServer) Serve() error {

	for {
		conn, err := srv.l.AcceptTCP()
		if err != nil {
			break
		}
		srv.currCount.Add(1)
		go srv.HandleConn(conn)
	}

	return nil
}

func (ch *CacheServer) currentConn() int64 {
	return ch.currCount.Load()
}

func (ch *CacheServer) HandleConn(conn *net.TCPConn) {

	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(1 * time.Second)
	// enc := protocol.NewEncoder(conn)
	dec := protocol.NewDecoder(conn)
	defer func() {
		conn.Close()
		ch.currCount.Add(-1)
		// ch.logger.Info("connection: ", "count", ch.currentConn())
	}()

	var args protocol.Array
	for {
		err := dec.Decode(&args)
		if err != nil {
			if err != io.EOF {
				ch.logger.Error(err.Error())
			}
			return
		}

		cmd, ok := args[0].(protocol.Bulk)
		if !ok {
			err = fmt.Errorf("expected %T type, got { %T }", protocol.Bulk{}, args[0])
			ch.logger.Error(err.Error())
			return
		}

		switch string(cmd) {
		case "hello", "HELLO":

			status := protocol.SimpleErr(fmt.Errorf("what?"))
			if err := protocol.WriteReply(conn, status); err != nil {
				ch.logger.Debug(err.Error())
				return
			}

		case "SET", "set":
			sa := protocol.SetARGS{}
			if err := sa.FromArgs(args[1:]); err != nil {
				// ch.logger.Error(err.Error())
				return
			}

			if err := ch.handleSET(conn, sa.Key(), sa.Value(), sa.Expire()); err != nil {
				ch.logger.Debug(err.Error())
				return
			}

		case "get", "GET":
			key, ok := args[1].(protocol.Bulk)
			if !ok {
				ch.logger.Error("not bulk type", "type", fmt.Errorf("%T", args[1]))
				return
			}

			err := ch.handleGET(conn, key)
			if err != nil {
				ch.logger.Debug(err.Error())
				return
			}

		default:
			protocol.WriteReply(conn, fmt.Errorf("unknown command {%s}", cmd))
			return
		}

	}

}

func (ch *CacheServer) handleGET(w io.Writer, key []byte) error {

	val, ok := ch.cache.Get(key)
	if !ok {
		return protocol.WriteReply(w, fmt.Errorf("not found"))
	}

	if err := protocol.WriteReply(w, val); err != nil {
		return err
	}
	return nil
}

func (ch *CacheServer) handleSET(w io.Writer, key []byte, value []byte, ttl int64) error {
	if err := ch.cache.Set(key, value, ttl); err != nil {
		return protocol.WriteReply(w, fmt.Errorf("err %v", err))
	}
	return protocol.WriteReply(w, "OK")
}

// func (ch *CacheServer) handleErr(w io.Writer, err error) {
// 	if err != nil {
// 		// if err == io.EOF {
// 		// 	return
// 		// }
// 		ch.logger.Debug(err.Error())
// 		protocol.WriteReply(w, err)
// 	}
// }
