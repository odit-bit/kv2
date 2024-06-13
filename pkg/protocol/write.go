package protocol

import (
	"fmt"
	"io"
)

//WRITE

func writeBulkString(w io.Writer, val Bulk) (int, error) {
	n, err := fmt.Fprintf(w, "$%d\r\n%s\r\n", len(val), val)
	if err != nil {
		return 0, err
	}
	return n, err
}

func writeSimpleString(w io.Writer, val any) (int, error) {
	return fmt.Fprintf(w, "+%s\r\n", val)
}

func writeSimpleError(w io.Writer, val any) (int, error) {
	return fmt.Fprintf(w, "-%s\r\n", val)
}

func writeInteger(w io.Writer, v int64) (int, error) {
	n, err := fmt.Fprintf(w, ":%d\r\n", v)
	if err != nil {
		return 0, err
	}
	return n, err

}

func Write(w io.Writer, v any) error {
	switch v := v.(type) {
	case Integer:
		if _, err := writeInteger(w, int64(v)); err != nil {
			return err
		}
	case Bulk:
		if _, err := writeBulkString(w, v); err != nil {
			return err
		}
	case Array:
		if err := writeArray(w, v); err != nil {
			return err
		}
	case Map:
		if err := writeMap(w, v); err != nil {
			return err
		}
	case Simple, string:
		//same as []byte
		if _, err := writeSimpleString(w, v); err != nil {
			return err
		}
	case SimpleErr:
		if _, err := writeSimpleError(w, v); err != nil {
			return err
		}

	default:
		return fmt.Errorf("invalid type: %T", v)
	}
	return nil
}

// wrap vals element as agg array type
func writeArray(w io.Writer, val Array) error {
	//write array header "*<len>\r\n"
	if _, err := fmt.Fprintf(w, "%s%d%s", string(TYPE_AGG_ARRAY), len(val), string(CRLF)); err != nil {
		return err
	}

	for _, v := range val {
		if err := Write(w, v); err != nil {
			return err
		}
	}
	return nil
}

// write maps aggregate type
func writeMap(w io.Writer, m Map) error {
	//write the header for agg type
	if _, err := fmt.Fprintf(w, "%s%d%s", string(TYPE_AGG_MAP), len(m), string(CRLF)); err != nil {
		return err
	}

	for k, v := range m {
		//key
		if err := Write(w, k); err != nil {
			return err
		}

		//value
		if err := Write(w, v); err != nil {
			return err
		}

	}

	return nil

}

// helper function to create reply message to client as
// simpleError, simpleString, bulk([]byte)
func WriteReply(w io.Writer, msg any) error {
	var err error
	switch msg := msg.(type) {
	case string:
		_, err = writeSimpleString(w, Simple(msg))
	case error:
		_, err = writeSimpleError(w, SimpleErr(msg))
	case []byte:
		_, err = writeBulkString(w, Bulk(msg))
	default:
		err = fmt.Errorf("not implemented reply type")
	}

	return err
}
