package x

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/VictoriaMetrics/fastcache"
)

type Cache struct {
	db *fastcache.Cache
	mu sync.Mutex
}

// size in byte if size < 32mb , minimum size 32mb
func NewFastcache(cap int) *Cache {
	fc := fastcache.New(cap)
	cache := Cache{
		db: fc,
		mu: sync.Mutex{},
	}
	return &cache
}

func (c *Cache) Status() *fastcache.Stats {
	stat := fastcache.Stats{}
	c.db.UpdateStats(&stat)
	return &stat
}

func (c *Cache) Close() error {
	c.db.Reset()
	return nil
}

// type item struct {
// 	Value      []byte
// 	TTL        time.Time
// 	IsExpiring bool
// }

// func (it *item) setExpired(dur time.Duration) {
// 	if dur <= 0 {
// 		it.IsExpiring = false
// 		return
// 	}

// 	it.IsExpiring = true
// 	it.TTL = time.Now().Add(dur)
// }

// func (it *item) isValid() bool {
// 	if it.IsExpiring {
// 		return time.Now().Before(it.TTL)
// 	}
// 	return true
// }

const prefixTTL = "ttl"
const prefixKey = "key"

// func (c *Cache) Set(key, value []byte, ttl int64) error {
// 	vBuf := bpool.Get()
// 	defer bpool.Put(vBuf)

// 	it := item{
// 		Value: value,
// 	}
// 	it.setExpired(time.Duration(ttl))
// 	if err := gob.NewEncoder(vBuf).Encode(it); err != nil {
// 		return err
// 	}

// 	c.db.SetBig(key, vBuf.Bytes())
// 	vBuf.Reset()
// 	return nil
// }

// func (c *Cache) Get(key []byte) ([]byte, bool) {

// 	b := c.db.GetBig(nil, key)
// 	if len(b) == 0 {
// 		return nil, false
// 	}

// 	buf := bytes.NewReader(b)
// 	it := item{}
// 	gob.NewDecoder(buf).Decode(&it)

// 	if it.isValid() {
// 		c.db.Del(key)
// 		return it.Value, true
// 	}
// 	return nil, false

// }

func (c *Cache) Set(key, value []byte, dur int64) error {

	//make new key
	valKey := append(key, []byte(prefixKey)...)

	// Set ttl and value
	if dur > int64(time.Millisecond) {
		ttlKey := append(key, []byte(prefixTTL)...)
		c.setTTL(ttlKey, dur)
	}

	c.db.SetBig(valKey, value)
	return nil
}

func (c *Cache) setTTL(ttlKey []byte, n int64) {
	if n > int64(1*time.Millisecond) {
		expire := time.Now().Add(time.Duration(n)).UnixNano()
		ttlB := fmt.Sprintf("%d", expire)
		c.db.Set(ttlKey, []byte(ttlB))
	}

	expire := time.Now().Add(time.Duration(n)).UnixNano()
	ttlB := fmt.Sprintf("%d", expire)
	c.db.Set(ttlKey, []byte(ttlB))
}

func (c *Cache) Get(key []byte) ([]byte, bool) {
	//make new key
	valKey := append([]byte{}, key...)
	valKey = append(valKey, []byte(prefixKey)...)

	//get ttl and value
	ttlKey := append([]byte{}, key...)
	ttlKey = append(ttlKey, []byte(prefixTTL)...)

	// log.Panic(string(valKey), string(ttlKey))

	ttl := c.getTTL(ttlKey)
	if ttl > 0 {
		if time.Now().UnixNano() > ttl {
			c.db.Del(valKey)
			return nil, false
		}

	}

	// get value
	valGet := c.db.GetBig(nil, valKey)
	if len(valGet) == 0 {
		return nil, false
	}

	return valGet, true
}

func (c *Cache) getTTL(ttlKey []byte) int64 {

	//get ttl
	ttlGet := c.db.Get(nil, ttlKey)
	if len(ttlGet) > 0 {
		//parse the ttl into int64l
		ttl, err := strconv.ParseInt(string(ttlGet), 10, 64)
		if err != nil {
			log.Fatal(string(ttlGet))
			return 0
		}
		return ttl
	}
	return 0

}

func (c *Cache) Delete(key []byte) error {
	c.db.Del(key)
	return nil
}
