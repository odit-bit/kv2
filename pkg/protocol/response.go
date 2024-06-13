package protocol

import (
	"fmt"
	"io"
)

// return ok or error
func ReadSETCmdResponse(r io.ByteReader) (string, error) {
	b, err := r.ReadByte()
	if err != nil {
		return "", err
	}
	switch b {
	case TYPE_SIMPLE_STRING:
		msg, err := readSimpleString(r)
		if err != nil {
			return "", err
		}
		return string(msg), nil
	case TYPE_SIMPLE_ERR:
		msg, err := readSimpleErr(r)
		if err != nil {
			return "", err
		}
		return "", msg

	default:
		return "", fmt.Errorf("invalid set response: %v", string(b))
	}
}

// return bulk or error
func ReadGetCMDResponse(r io.ByteReader) ([]byte, error) {
	b, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	switch b {
	case TYPE_BULK:
		return readBulkString(r)

	case TYPE_SIMPLE_ERR:
		b, err := readSimpleErr(r)
		if err != nil {
			return nil, err
		}
		return nil, b
	default:
		return nil, fmt.Errorf("invalid get response: %v", string(b))
	}
}
