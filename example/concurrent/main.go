package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/odit-bit/kv2/client"
)

//use build-in client package. see /client

var dataSize = 64 << 10

var globVal = func() string {
	// fmt.Println("data size:", dataSize)
	value := strings.Repeat("X", dataSize)
	return value
}()

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	cli := client.New("localhost:6969")
	defer cli.Close()

	key := "myKey"
	value := globVal
	max := 200
	count := atomic.Int64{}
	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < max; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("%s-%d", key, i)
			value := fmt.Sprintf("%s-%d", globVal, i)

			if _, err := cli.Set([]byte(key), []byte(value), &client.SetOpt{}); err != nil {
				slog.Error("set:", "err", err)
				return
			}
			count.Add(1)
		}(i)

	}
	wg.Wait()

	for i := 0; i < max; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("%s-%d", key, i)
			expect := fmt.Sprintf("%s-%d", globVal, i)

			result, err := cli.Get([]byte(key))
			if err != nil {
				return
			}

			if len(result) != len(expect) {
				fmt.Println("get result not equal", len(value), len(result))
			}
			// //check every element
			// if !bytes.Equal(resp.Value, []byte(expect)) {
			// 	fmt.Println("get result not equal", len(value), len(resp.Value))
			// }
		}(i)
	}

	wg.Wait()
	end := time.Since(start)
	fmt.Println("count:", count.Load())
	fmt.Println("time:", end.String())

}
