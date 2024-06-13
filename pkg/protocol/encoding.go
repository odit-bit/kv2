package protocol

import (
	"bufio"
	"fmt"
	"io"
)

type ErrInvalidType struct {
	msg string
}

func (err *ErrInvalidType) Error() string {
	return err.msg
}

type Decoder struct {
	r *bufio.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	rr, ok := r.(*bufio.Reader)
	if !ok {
		rr = bufio.NewReader(r)
	}

	dec := Decoder{
		r: rr,
	}

	return &dec
}

func (dec *Decoder) Decode(v any) error {
	switch v := v.(type) {
	case *Bulk:
		b, err := parseBulk(dec.r)
		if err != nil {
			return err
		}
		*v = b
		return nil

	case *Integer:
		b, err := parseInteger(dec.r)
		if err != nil {
			return err
		}
		*v = b
		return nil

		//
	case *Simple:
		n, err := parseSimpleString(dec.r)
		if err != nil {
			return err
		}
		*v = n
		return nil
	case *Array:

		arr, err := parseArray(dec.r)
		if err != nil {
			return err
		}
		*v = arr

		return nil
	default:
		return fmt.Errorf("invalid type parameter {%T} ", v)
	}
}

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: w,
	}
}

func (enc *Encoder) Encode(v any) error {
	return Write(enc.w, v)
}

//////////////////////////////////

// type CommandReader struct {
// 	br        *bufio.Reader
// 	argsCount int
// }

// func NewCmdReader(r io.Reader) (*CommandReader, error) {
// 	br, ok := r.(*bufio.Reader)
// 	if !ok {
// 		br = bufio.NewReader(r)
// 	}

// 	b, err := br.Peek(1)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if b[0] != TYPE_AGG_ARRAY {
// 		return nil, fmt.Errorf("malformed cmd packet, [ %s ]", string(b[0]))
// 	}
// 	_, _ = br.Discard(1)

// 	//read argsCount (array length)
// 	n, err := readInteger(br)
// 	if err != nil {
// 		return nil, err
// 	}
// 	cr := CommandReader{
// 		br:        br,
// 		argsCount: int(n),
// 	}
// 	return &cr, nil
// }

// func (cr *CommandReader) NextArg() (any, error) {
// 	if cr.argsCount > 0 {
// 		return parseElement(cr.br)
// 	}
// 	return nil, fmt.Errorf("no more args")
// }

// func (cr *CommandReader) WriteNext(w io.Writer) error {
// 	cr.argsCount--
// }

// func (cr *CommandReader) read() (any, error) {
// 	cr.argsCount--
// 	return parseElement(cr.br)
// }
