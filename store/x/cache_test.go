package x

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/VictoriaMetrics/fastcache"
	"github.com/dustin/go-humanize"
	"github.com/stretchr/testify/assert"
)

func Test_Set_Get(t *testing.T) {
	key := []byte("key")
	value := bytes.Repeat([]byte{'x'}, 128<<10)

	fc := fastcache.New(128 << 20)
	cache := Cache{
		db: fc,
	}

	//set get
	if err := cache.Set(key, value, 0); err != nil {
		t.Fatal(err)
	}
	res, ok := cache.Get(key)
	if !ok {
		t.Fatal("value should exist")
	}
	assert.Equal(t, value, res)

	// set get expired
	if err := cache.Set(key, value, int64(10*time.Millisecond)); err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)

	_, ok = cache.Get(key)
	if ok {
		t.Fatalf("get expired key should return false")
	}
}

func Test_Concurrent(t *testing.T) {
	fc := fastcache.New(100000000)
	cache := Cache{
		db: fc,
	}
	defer fc.Reset()

	key := []byte("key")
	value := bytes.Repeat([]byte{'x'}, 10000)

	var wg sync.WaitGroup
	numGoRoutines := 200

	type test struct {
		key   []byte
		value []byte
		ttl   int64
	}

	tTable := make([]test, numGoRoutines)
	for i := range tTable {
		keyB := fmt.Sprintf("%s-%d", key, i)
		valB := fmt.Sprintf("%s-%d", value, i)
		tTable[i] = test{
			key:   []byte(keyB),
			value: []byte(valB),
			ttl:   0,
		}
	}

	//concurrent write
	for _, tc := range tTable {
		wg.Add(1)
		go func(tc *test) {
			defer wg.Done()
			if err := cache.Set(tc.key, tc.value, tc.ttl); err != nil {
				t.Fail()
				t.Log(err)
				return
			}
		}(&tc)
	}

	wg.Wait()

	// concurrent read
	for _, tc := range tTable {
		wg.Add(1)
		go func(tc *test) {
			defer wg.Done()
			b, ok := cache.Get(tc.key)
			if !ok {
				t.Fail()
				t.Logf("value of %s not existed", string(tc.key))
				return
			}

			assert.Equal(t, tc.value, b)
		}(&tc)
	}

	wg.Wait()

}

func Test_buffer(t *testing.T) {
	c := NewFastcache(10 * humanize.MByte)

	//store
	// 1st store the ttl
	// 2nd store the value
	// while storing it will allocate
	prefixKey := []byte("key")
	prefixTTL := []byte("ttl")
	key := append([]byte("key"), prefixKey...)
	ttlKey := append([]byte("key"), prefixTTL...)
	value := []byte("value")
	ttl := []byte(strconv.Itoa(int(64)))

	//store ttl
	c.db.Set(ttlKey, ttl)
	//store key
	c.db.SetBig(key, value)

	//get ttl
	// buffer := make([]byte, 0, 10000)
	ttlGet := c.db.Get(nil, ttlKey)
	assert.Equal(t, ttl, ttlGet)

	//get key
	valueGet := c.db.GetBig(nil, key)
	assert.Equal(t, value, valueGet)
}

func Test_time_format(t *testing.T) {
	dur := 10 * time.Millisecond
	now := time.Now()
	expect := []byte(fmt.Sprintf("%d", now.Add(dur).UnixNano()))

	actual := []byte(strconv.FormatInt(now.Add(dur).UnixNano(), 10))
	assert.Equal(t, expect, actual)

	res, err := strconv.ParseInt(string(expect), 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, dur.Nanoseconds(), res-now.UnixNano())

	// convert 0
	res, err = strconv.ParseInt(string([]byte{}), 10, 0)
	if err == nil {
		t.Fatal("cannot parse '' ")
	}
	assert.Equal(t, int64(0), res)
}

// benchmark
var stat = fastcache.Stats{}

func Benchmark_(b *testing.B) {
	b.StopTimer()

	cache := NewFastcache(6400000)

	tMap := map[string][]byte{}
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key-%d", i)
		val := make([]byte, 64000)
		rand.Read(val)
		tMap[key] = val
	}

	b.StartTimer()
	for k, v := range tMap {
		if err := cache.Set([]byte(k), v, 0); err != nil {
			b.Fatal()
		}
	}

	cache.db.UpdateStats(&stat)
	// for i := 0; i < b.N; i++ {
	// 	if err := cache.Set([]byte("key"), val, 0); err != nil {
	// 		b.Fatal()
	// 	}
	// }
}
