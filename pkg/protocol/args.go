package protocol

import (
	"fmt"
	"time"
)

type SetARGS struct {
	// CMD      string
	key      []byte
	value    []byte
	timeUnit string
	expire   int64
}

func (sc *SetARGS) Key() []byte {
	return sc.key
}

func (sc *SetARGS) Value() []byte {
	return sc.value
}

func (sc *SetARGS) Expire() int64 {
	return int64(sc.toDuration())
}

// calculate duration of expire time according to TimeUnit
func (ttl *SetARGS) toDuration() time.Duration {
	switch ttl.timeUnit {
	case "EX":
		//value is second need convert to nanosecond
		return time.Duration(ttl.expire * 1000000000)
	case "PX":
		//value is milisecond need convert to nanosecond
		return time.Duration(ttl.expire * 1000000)
	case "EXAT":
		// value is unix-time in second need convert to nanosecond
		return time.Duration(time.Unix(ttl.expire, 0).Unix())
	case "PXAT":
		// value is unix-time in milliseocnd need convert to nanosecond
		return time.Duration(time.UnixMilli(ttl.expire).Unix())

	default:
		// is nanosecond
		// should return error instead ?
		return time.Duration(ttl.expire)
	}
}

func (ch *SetARGS) FromArgs(args []any) error {
	return ch.fromArgs(args)
}

func (ch *SetARGS) fromArgs(arr []any) error {
	var ok bool
	//KEY
	ch.key, ok = arr[0].(Bulk)
	if !ok {

		return ch.handleError(fmt.Sprintf("expect Key type is bytes/bulkstring got :%t", arr[0]))
	}

	//VALUE
	ch.value, ok = arr[1].(Bulk)
	if !ok {
		return ch.handleError(fmt.Sprintf("expect Value type is bytes/bulkstring got :%t", arr[1]))
	}

	if len(arr) == 2 {
		return nil
	}

	//TIME TO LIVE ARGS
	//EX ,PX,EXAT,PXAT
	unit, ok := arr[2].(Bulk)
	if !ok {
		return ch.handleError(fmt.Sprintf("expect time unit type is bulk got :%t", arr[2]))
	}
	ch.timeUnit = string(unit)
	if ok := ch.checkTimeUnit(); !ok {
		return ch.handleError(fmt.Sprintf("wrong time unit {%v}", string(unit)))
	}

	exp, ok := arr[3].(Integer)
	if !ok {
		return ch.handleError(fmt.Sprintf("expect expire time type is integer got :%t", arr[3]))
	}
	ch.expire = int64(exp)
	return nil

}

func (sc *SetARGS) checkTimeUnit() bool {
	switch sc.timeUnit {
	case "EX", "PX", "EXAT", "PXAT":
		return true
	default:
		return false
	}
}

func (sc *SetARGS) handleError(msg any) error {
	return fmt.Errorf("set command protocol: %v", msg)
}

type GetARGS struct {
	Key []byte
}

func (get *GetARGS) fromArgs(arr []any) error {
	if len(arr) != 1 {
		return fmt.Errorf("wrong args %v", len(arr))
	}

	v, ok := arr[0].([]byte)
	if !ok {
		v, ok := arr[0].(string)
		if !ok {
			return fmt.Errorf("get: key args wrong type [%T]", arr[0])
		}
		get.Key = []byte(v)
	} else {
		get.Key = v
	}

	return nil
}
