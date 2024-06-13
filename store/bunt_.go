package store

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/tidwall/buntdb"
)

type MemCache struct {
	db *buntdb.DB
}

func NewInMem() *MemCache {
	bdb, err := buntdb.Open(":memory:")
	if err != nil {
		log.Fatal()
	}

	rb := MemCache{
		db: bdb,
	}

	return &rb
}

func (r *MemCache) Close() error {
	return r.db.Close()
}

// Delete implements Storer.
func (r *MemCache) Delete(key []byte) error {
	return fmt.Errorf("un implemented")
}

// Get implements Storer.
func (r *MemCache) Get(key []byte) ([]byte, bool) {
	var buf bytes.Buffer
	err := r.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(string(key), false)
		if err != nil {
			return err
		}
		if _, err := buf.WriteString(val); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, false
	}
	return buf.Bytes(), true
}

// Set implements Storer.
func (bunt *MemCache) Set(key []byte, value []byte, ttl int64) error {
	var exp bool
	if ttl != 0 {
		exp = true
	}


	return bunt.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(string(key), string(value), &buntdb.SetOptions{
			Expires: exp,
			TTL:     time.Duration(ttl),
		})
		return err
	})

}

// encapsulate pair data , use for restore and persist
type data struct {
	Key   []byte
	Value []byte
	TTL   int64
}

// it will drop all the old existed data, restore from r.
// the data from r should as
//
/*
	map[string][]byte
*/
// it is txn if it error it will rolled back
func (rb *MemCache) Restore(r io.Reader) error {
	//drop the data

	dec := json.NewDecoder(r)
	data := data{}
	err := rb.db.Update(func(tx *buntdb.Tx) error {
		if err := tx.DeleteAll(); err != nil {
			return nil
		}
		var uErr error
		for {
			if err := dec.Decode(&data); err == io.EOF {
				break
			}

			isTTL := false
			if data.TTL > 0 {
				isTTL = true
			}
			_, _, err := tx.Set(string(data.Key), string(data.Value), &buntdb.SetOptions{
				TTL:     time.Duration(data.TTL),
				Expires: isTTL,
			})
			if err != nil {
				uErr = err
				break
			}

		}
		return uErr
	})
	if err != nil {
		return err
	}
	return nil

}

// dump all existed data into w with data formated as json,
/*
	map[[]byte][]byte
*/
// it is txn if it error it will rolled back
func (rb *MemCache) Persist(w io.Writer) error {
	bw, ok := w.(*bufio.Writer)
	if !ok {
		bw = bufio.NewWriter(w)
	}

	err := rb.db.View(func(tx *buntdb.Tx) error {
		var vErr error
		err := tx.Ascend("", func(key, value string) bool {
			vErr = nil
			ttl, err := tx.TTL(key)
			if err != nil {
				vErr = err
				return false
			}

			pair := data{
				Key:   []byte(key),
				Value: []byte(value),
				TTL:   int64(ttl),
			}

			b, err := json.Marshal(pair)
			if err != nil {
				vErr = err
				return false
			}
			if _, err := bw.Write(b); err != nil {
				vErr = err
				return false
			}
			if _, err := bw.Write([]byte{'\n'}); err != nil {
				vErr = err
				return false
			}

			return true
		})
		//error for writing
		if vErr != nil {
			return vErr
		}
		// error for iterate
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	if err := bw.Flush(); err != nil {
		return err
	}
	return nil
}
