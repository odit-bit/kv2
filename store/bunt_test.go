package store

import (
	"bytes"
	"crypto/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tidwall/buntdb"
)

func Test_Ops(t *testing.T) {
	bdb, err := buntdb.Open(":memory:")
	if err != nil {
		t.Fatal()
	}
	defer bdb.Close()

	rb := MemCache{
		db: bdb,
	}

	type data struct {
		Key   []byte
		Value []byte
		TTL   time.Duration
	}
	d := data{
		Key:   []byte("key"),
		Value: []byte("value"),
		TTL:   0,
	}
	d2 := data{
		Key:   []byte("key2"),
		Value: []byte("value2"),
		TTL:   0,
	}

	//SET
	if err := rb.Set(d.Key, d.Value, int64(d.TTL)); err != nil {
		t.Fatal(err)
	}

	//SET2
	if err := rb.Set(d2.Key, d2.Value, int64(d2.TTL)); err != nil {
		t.Fatal(err)
	}

	//PERSIST
	backend := bytes.Buffer{}
	if err := rb.Persist(&backend); err != nil {
		t.Fatal(err)
	}

	//RESTORE
	bdb2, err := buntdb.Open(":memory:")
	if err != nil {
		t.Fatal()
	}
	defer bdb2.Close()
	rb2 := MemCache{
		db: bdb,
	}

	if err := rb2.Restore(&backend); err != nil {
		t.Fatal(err)
	}

	//GET
	v, ok := rb2.Get(d.Key)
	if !ok {
		t.Fatal("rb2 value should exist")
	}

	assert.Equal(t, d.Value, v)

}

//// benchmark

func Benchmark_(b *testing.B) {
	cache := NewInMem()
	val := make([]byte, 64024)
	rand.Read(val)

	for i := 0; i < b.N; i++ {
		if err := cache.Set([]byte("key"), val, 0); err != nil {
			b.Fatal()
		}
	}
}
