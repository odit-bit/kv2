package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/redis/go-redis/v9"
)

var (
	dataSize int
	max      int
)

func genValue(dataSize int) string {
	// fmt.Println("data size:", dataSize)
	value := strings.Repeat("X", dataSize)
	return value
}

func init() {

	flag.IntVar(&max, "max", 10, "")
	flag.IntVar(&dataSize, "size", 64, "")
	flag.Parse()

	dataSize = dataSize << 10

}

func main() {

	cli := redis.NewClient(&redis.Options{
		Addr:             "localhost:6969",
		DisableIndentity: true,
	})
	defer cli.Close()
	var key = "myKey"
	var total = (dataSize * max) + (len(key) * max)
	fmt.Println("data size: ", humanize.Bytes(uint64(dataSize)))
	fmt.Println("total size: ", humanize.Bytes(uint64(total)))

	type tCase struct {
		key   string
		value []byte
	}

	globVal := genValue(dataSize)
	tTable := make([]*tCase, max)
	for i := range tTable {
		tkey := fmt.Sprintf("%s-%d", key, i)
		tvalue := fmt.Sprintf("%s-%d", globVal, i)

		tTable[i] = &tCase{
			key:   tkey,
			value: []byte(tvalue),
		}
	}

	// var count = atomic.Int64{}
	var start = time.Now()
	var wg sync.WaitGroup

	for _, test := range tTable {
		wg.Add(1)
		go func(test *tCase) {
			defer wg.Done()
			res := cli.Set(context.Background(), test.key, test.value, 0)
			if res.Err() != nil {
				log.Printf("SET: %v, %v \n", res.Err(), test.key)
				return
			}
		}(test)

	}

	wg.Wait()
	fmt.Println("set done: ", time.Since(start))

	start = time.Now()
	for _, test := range tTable {
		wg.Add(1)
		go func(test *tCase) {
			defer wg.Done()

			res := cli.Get(context.Background(), test.key)
			if res.Err() != nil {
				log.Fatalf("GET: %v, %v \n", res.Err(), test.key)
			}
			actual, err := res.Bytes()
			if err != nil {
				log.Fatal(res.Err(), test.key)
			}
			if len(actual) != len(test.value) {
				fmt.Println("get result not equal", len(actual), len(test.value))
			}

		}(test)
	}

	wg.Wait()
	fmt.Println("get done: ", time.Since(start))

}
