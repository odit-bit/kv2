package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/odit-bit/kv2/pkg/protocol"
	"github.com/odit-bit/kv2/store/x"
	"github.com/panjf2000/gnet/v2"
)

type App struct {
	handler *Handler
	logger  *slog.Logger
}

func NewEdge() *App {
	cacher := x.NewFastcache(512000000)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}))
	handler := Handler{
		BuiltinEventEngine: gnet.BuiltinEventEngine{},
		logger:             logger,
		cache:              cacher,
	}

	app := App{
		handler: &handler,
		logger:  logger,
	}
	return &app
}

func (app *App) Run() error {
	var wg sync.WaitGroup
	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(1)
	go func() {
		<-sigC
		app.logger.Info("got signal")
		if err := app.handler.eng.Stop(context.Background()); err != nil {
			app.logger.Error("shutting down server", "status", err)
		}
		wg.Done()
	}()

	addr := "tcp://:6969"
	err := gnet.Run(app.handler,
		addr,
		gnet.WithMulticore(true),
		gnet.WithEdgeTriggeredIO(true),
		// gnet.WithReadBufferCap(rBufCap),
		// gnet.WithWriteBufferCap(wBufCap),
	)

	if err != nil {
		app.logger.Error(err.Error())
	}

	wg.Wait()
	time.Sleep(100 * time.Millisecond)
	return err
}

///Handler

var _ gnet.EventHandler = (*Handler)(nil)

type Handler struct {
	gnet.BuiltinEventEngine
	eng       gnet.Engine
	logger    *slog.Logger
	cache     Cacher
	currCount atomic.Int64
}

func NewHandler(cache Cacher) *Handler {
	h := Handler{
		BuiltinEventEngine: gnet.BuiltinEventEngine{},
		cache:              cache,
	}
	return &h
}

func (ch *Handler) OnBoot(e gnet.Engine) gnet.Action {
	ch.logger.Info("ready receive connection")
	ch.eng = e
	return gnet.None

}

func (ch *Handler) OnClose(c gnet.Conn, err error) gnet.Action {
	if err != nil {
		if err != io.EOF {
			ch.logger.Info("error occurred on connection", "addr", c.RemoteAddr().String(), "msg", err)
		}
	}
	ch.currCount.Add(-1)
	return gnet.None
}

func (ch *Handler) OnOpen(c gnet.Conn) ([]byte, gnet.Action) {
	ch.currCount.Add(1)
	return nil, gnet.None
}

func (ch *Handler) OnTraffic(c gnet.Conn) gnet.Action {

	buf := bytes.NewBuffer([]byte{})
	for {
		b, err := c.Peek(1)
		if err != nil {
			log.Println(err)
			break
		}

		if _, err := buf.Write(b); err != nil {
			log.Println(err)
			return gnet.Close
		}
		time.Sleep(100 * time.Millisecond)

	}
	fmt.Println("packet size", humanize.Bytes(uint64(buf.Len())))

	// buf := bytes.Buffer{}
	// args, err := command.XReadNextCommand(c, &buf)
	// if err != nil {
	// 	if err != io.EOF {
	// 		ch.logger.Error(err.Error())
	// 	}
	// 	return gnet.Close
	// }

	// cmd := string(args[0])
	// switch cmd {
	// case "SET", "set":
	// 	var ttl int64
	// 	if len(args) >= 5 {
	// 		ttlMode := args[3]
	// 		ttlVal := args[4]
	// 		ttl = command.ExToDuration(string(ttlMode), ttlVal)
	// 	}
	// 	if err := ch.handleSET(c, args[1], args[2], ttl); err != nil {
	// 		// if err != io.EOF {
	// 		// 	ch.logger.Error(err.Error())
	// 		// }
	// 		return gnet.Close
	// 	}

	// case "GET", "get":
	// 	if err := ch.handleGET(c, args[1]); err != nil {
	// 		// if err != io.EOF {
	// 		// 	ch.logger.Error(err.Error())
	// 		// }
	// 		return gnet.Close
	// 	}

	// default:
	// 	err := fmt.Errorf("command not implemented %v", cmd)
	// 	protocol.Write(c, err)
	// 	return gnet.Close
	// }

	return gnet.Close
}

// func (ch *Handler) OnTraffic(c gnet.Conn) gnet.Action {
// 	var none = gnet.None
// 	var close = gnet.Close

// 	if err := c.SetReadBuffer(64 << 20); err != nil {
// 		ch.logger.Error(err.Error())
// 		return close
// 	}
// 	dec := protocol.NewDecoder(c)
// 	enc := c

// 	var args protocol.Array
// 	err := dec.Decode(&args)
// 	if err != nil {
// 		if err := protocol.WriteReply(enc, err); err != nil {
// 			log.Fatal(err)
// 			ch.logger.Error(err.Error())
// 		}
// 		return close
// 	}

// 	cmd, ok := args[0].(protocol.Bulk)
// 	if !ok {
// 		err = fmt.Errorf("expected %T type, got { %T }", protocol.Bulk{}, args[0])
// 		if err := protocol.WriteReply(enc, err); err != nil {
// 			ch.logger.Error(err.Error())
// 		}
// 		return close
// 	}

// 	switch string(cmd) {
// 	case "hello", "HELLO":
// 		status := protocol.SimpleErr(fmt.Errorf("what?"))
// 		err := protocol.WriteReply(enc, status)
// 		if err != nil {
// 			ch.logger.Debug(err.Error())
// 			return close
// 		}

// 	case "SET", "set":
// 		sa := command.SetARGS{}
// 		if err := sa.FromArgs(args[1:]); err != nil {
// 			ch.logger.Debug(err.Error())
// 			protocol.WriteReply(enc, err)
// 			return close
// 		}

// 		if err := ch.handleSET(enc, sa.Key(), sa.Value(), sa.Expire()); err != nil {
// 			ch.logger.Debug(err.Error())
// 			return close
// 		}

// 	case "get", "GET":
// 		key, ok := args[1].(protocol.Bulk)
// 		if !ok {
// 			if err := protocol.WriteReply(enc, fmt.Errorf("not bulk type: %T", args[1])); err != nil {
// 				ch.logger.Debug(err.Error())
// 			}
// 			return close
// 		}

// 		if err := ch.handleGET(enc, key); err != nil {
// 			ch.logger.Debug(err.Error())
// 			return close
// 		}

// 	default:
// 		err := fmt.Errorf("unknown command {%s}", cmd)
// 		if err := protocol.WriteReply(enc, err); err != nil {
// 			ch.logger.Error(err.Error())
// 		}
// 		return close
// 	}

// 	// c.Write(enc.Bytes())
// 	return none

// }

func (ch *Handler) handleGET(w io.Writer, key []byte) error {

	val, ok := ch.cache.Get(key)
	if !ok {
		return protocol.WriteReply(w, fmt.Errorf("not found"))
	}

	if err := protocol.WriteReply(w, val); err != nil {
		return err
	}
	return nil
}

func (ch *Handler) handleSET(w io.Writer, key []byte, value []byte, ttl int64) error {
	if err := ch.cache.Set(key, value, ttl); err != nil {
		return protocol.WriteReply(w, fmt.Errorf("err %v", err))
	}
	return protocol.WriteReply(w, "OK")
}
