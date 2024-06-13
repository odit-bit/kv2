package protocol

import (
	"errors"
	"fmt"
	"io"
	"strconv"
)

func Read(r io.ByteReader) (any, error) {
	return parseElement(r)
}
func IsValidType(r io.ByteReader, t byte) error {
	b, err := r.ReadByte()
	if err != nil {
		return err
	}
	if b != t {
		return fmt.Errorf("invalid type, expected (%c) got [%c]", t, b)
	}
	return nil
}

func parseArray(r io.ByteReader) (Array, error) {
	if err := IsValidType(r, TYPE_AGG_ARRAY); err != nil {
		return nil, err
	}

	size, err := ReadLength(r)
	if err != nil {
		return nil, err
	}

	arr := make([]any, size)
	for i := range size {
		el, err := parseElement(r)
		if err != nil {
			return nil, err
		}
		arr[i] = el
	}
	return arr, nil
}

func parseElement(r io.ByteReader) (any, error) {
	b, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	var v any
	switch b {
	case TYPE_INTEGER:
		v, err = readInteger(r)

	case TYPE_BULK: //[]byte
		v, err = readBulkString(r)

	case TYPE_SIMPLE_ERR:
		v, err = readSimpleErr(r)

	case TYPE_SIMPLE_STRING:
		v, err = readSimpleString(r)

	case TYPE_AGG_MAP:
		err = fmt.Errorf("not implemented ")

	case TYPE_AGG_ARRAY:
		v, err = parseArray(r)

	default:
		return nil, fmt.Errorf("protocol: unknown element type : %s", string(b))

	}

	if err != nil {
		return nil, err
	}
	return v, nil
}

func parseSimpleString(r io.ByteReader) (Simple, error) {
	if err := IsValidType(r, TYPE_SIMPLE_STRING); err != nil {
		return "", err
	}

	str, err := readSimpleString(r)
	if err != nil {
		return "", err
	}

	res := Simple(str)
	return res, nil
}

func readSimpleString(r io.ByteReader) ([]byte, error) {
	simple, err := readUntilCR(r)
	if err != nil {
		return nil, err
	}

	return simple, err
}

func readSimpleErr(r io.ByteReader) (SimpleErr, error) {
	simple, err := readUntilCR(r)
	if err != nil {
		return nil, err
	}

	return errors.New(string(simple)), err
}

func parseInteger(r io.ByteReader) (Integer, error) {
	if err := IsValidType(r, TYPE_INTEGER); err != nil {
		return 0, err
	}
	return readInteger(r)
}

func readInteger(r io.ByteReader) (Integer, error) {
	b, err := readUntilCR(r)
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(string(b))
	if err != nil {
		return 0, err
	}
	return Integer(n), nil
}

func parseBulk(r io.ByteReader) (Bulk, error) {
	//read type
	if err := IsValidType(r, TYPE_BULK); err != nil {
		return nil, err
	}

	return readBulkString(r)
}

func readBulkString(r io.ByteReader) (Bulk, error) {

	//get length of bulkstring
	length, err := ReadLength(r)
	if err != nil {
		return nil, err
	}

	// get data of bulk
	data, err := readUntilCR(r) //r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	if len(data) != length {
		return nil, fmt.Errorf("read bulkString data [%v] not match with length [%v]", length, len(data))
	}

	return data, nil

}

// read length of message
func ReadLength(r io.ByteReader) (int, error) {
	//get the length of message
	lengthB, err := readUntilCR(r) //r.ReadBytes('\n')
	if err != nil {
		return 0, err
	}

	//switch lengthB into int
	length, err := strconv.Atoi(string(lengthB))
	if err != nil {
		return 0, err
	}
	return length, nil
}

// cr not included
func readUntilCR(r io.ByteReader) ([]byte, error) {
	bytes := []byte{}
	// var gotR bool
	// for {
	// 	b, err := r.ReadByte()
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	switch b {
	// 	case '\r':
	// 		gotR = true

	// 	case '\n':
	// 		if gotR {
	// 			return bytes[:len(bytes)-1], nil
	// 		}
	// 	default:
	// 		gotR = false
	// 	}

	// 	bytes = append(bytes, b)
	// }

	for {
		b, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		switch b {
		case '\r':
			continue
		case '\n':
			return bytes, nil
		default:
			bytes = append(bytes, b)
		}
	}
}
