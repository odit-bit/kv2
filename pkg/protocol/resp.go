package protocol

import (
	"io"
)

const (
	TYPE_BULK          byte = '$'
	TYPE_INTEGER       byte = ':'
	TYPE_SIMPLE_STRING byte = '+'
	TYPE_SIMPLE_ERR    byte = '-'

	TYPE_AGG_ARRAY byte = '*'
	TYPE_AGG_MAP   byte = '%'
)

const (
	CRLF = "\r\n"
)

type Integer int64

// func (i *Integer) WriteTo(w io.Writer) (int64, error) {
// 	n, err := fmt.Fprintf(w, ":%d\r\n", i)
// 	if err != nil {
// 		return 0, err
// 	}
// 	return int64(n), err
// }

// func (i *Integer) ReadFrom(r io.Reader) (int64, error) {
// 	rr, ok := r.(*bufio.Reader)
// 	if !ok {
// 		rr = bufio.NewReader(r)
// 	}
// 	n, err := parseInteger(rr)
// 	if err != nil {
// 		return 0, err
// 	}
// 	return int64(n), nil
// }

type Bulk []byte

// func (bulk *Bulk) WriteTo(w io.Writer) (int64, error) {
// 	n, err := writeBulkString(w, *bulk)
// 	return int64(n), err
// }

// func (i *Bulk) ReadFrom(r io.Reader) (int64, error) {
// 	rr, ok := r.(*bufio.Reader)
// 	if !ok {
// 		rr = bufio.NewReader(r)
// 	}

// 	v, err := parseBulk(rr)
// 	if err != nil {
// 		return 0, err
// 	}

// 	*i = v
// 	return 0, err
// }

type Simple string
type SimpleErr error

type Array []any
type Map map[any]any

type RespType interface {
	Encode(w io.Writer) error
	Decode(r io.Reader) error
}
