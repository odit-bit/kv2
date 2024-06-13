package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/odit-bit/kv2/client"
)

//use build-in client package. see /client

var (
	_DataSize int
	_Max      int
)

func genValue(dataSize int) string {
	value := strings.Repeat("X", dataSize)
	return value
}

func init() {

	flag.IntVar(&_Max, "max", 100, "")
	flag.IntVar(&_DataSize, "size", 1024, "")
	flag.Parse()
	_DataSize = _DataSize << 10
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	cli := client.New("localhost:6969")
	defer cli.Close()
	var wg sync.WaitGroup

	var key = "myKey"
	var total = (_DataSize * _Max) + (len(key) * _Max)
	globVal := genValue(_DataSize)

	fmt.Println("data size: ", humanize.Bytes(uint64(_DataSize)))
	fmt.Println("total size: ", humanize.Bytes(uint64(total)))

	count := atomic.Int64{}
	start := time.Now()

	for i := 0; i < _Max; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("mykey-%d", i)
			value := fmt.Sprintf("%s-%d", globVal, i)

			if _, err := cli.Set([]byte(key), []byte(value), &client.SetOpt{}); err != nil {
				slog.Error("set:", "err", err)
				return
			}

			result, err := cli.Get([]byte(key))
			if err != nil {
				log.Println(err)
				return
			}
			if !bytes.Equal(result, []byte(value)) {
				fmt.Println("get result not equal", i, len(value), len(result))
			}
			count.Add(1)
		}(i)

	}
	wg.Wait()

	end := time.Since(start)
	fmt.Println("count:", count.Load())
	fmt.Println("time:", end.String())

}
