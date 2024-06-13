package protocol

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_WriteArray(t *testing.T) {
	expect := "*2\r\n$3\r\nGET\r\n$5\r\nmyKey\r\n"
	network := bytes.Buffer{}
	Write(&network, []any{Bulk("GET"), Bulk("myKey")})
	assert.Equal(t, expect, network.String())
}

func Test_WriteSETcmd(t *testing.T) {
	buf := bytes.Buffer{}
	err := Write(&buf, []any{Bulk("SET"), Bulk("key"), Bulk("value"), Bulk("EX"), Integer(5)})
	if err != nil {
		t.Fatal(err)
	}

	expect := "*5\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n$2\r\nEX\r\n:5\r\n"
	assert.Equal(t, expect, buf.String())
}
