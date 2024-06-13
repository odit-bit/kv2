package protocol

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_decoder(t *testing.T) {
	type test struct {
		Test   string
		input  string
		output any
		Expect any
		err    error
	}

	tTable := []test{
		{
			Test:   "simpleString",
			input:  "+OK\r\n",
			output: new(Simple),
			Expect: Simple("OK"),
			err:    nil,
		},
		{
			Test:   "Integer",
			input:  ":5\r\n",
			output: new(Integer),
			Expect: Integer(5),
			err:    nil,
		},
		{
			Test:   "bulk",
			input:  "$3\r\nSET\r\n",
			output: new(Bulk),
			Expect: Bulk("SET"),
			err:    nil,
		},

		{
			Test:   "array",
			input:  "*0\r\n",
			output: new(Array),
			Expect: Array{},
			err:    nil,
		},
	}

	buf := bytes.Buffer{}
	dec := NewDecoder(&buf)
	for _, tc := range tTable {
		buf.WriteString(tc.input)
		err := dec.Decode(tc.output)
		if assert.Equal(t, tc.err, err) {
			switch act := tc.output.(type) {
			case *Simple:
				expect := tc.Expect.(Simple)
				assert.Equal(t, expect, *act)
			case *Integer:
				expect := tc.Expect.(Integer)
				assert.Equal(t, expect, *act)
			case *Bulk:
				expect := tc.Expect.(Bulk)
				assert.Equal(t, expect, *act)
			case *Array:
				expect := tc.Expect.(Array)
				assert.Equal(t, expect, *act)
			default:
				t.Fatalf("got actual %T, expect %T", act, tc.Expect)
			}
		}

	}
}

func Test_encode(t *testing.T) {

	type test struct {
		Name   string
		Input  any
		Expect []byte
		err    error
	}

	tTable := []test{
		{
			Name:   "integer",
			Input:  Integer(10),
			Expect: []byte(":10\r\n"),
			err:    nil,
		},
		{
			Name:   "bulk",
			Input:  Bulk("SET"),
			Expect: []byte("$3\r\nSET\r\n"),
			err:    nil,
		},
		{
			Name:   "simple",
			Input:  Simple("OK"),
			Expect: []byte("+OK\r\n"),
			err:    nil,
		},

		{
			Name:   "simpleErr",
			Input:  SimpleErr(errors.New("ERROR")),
			Expect: []byte("-ERROR\r\n"),
			err:    nil,
		},

		{
			Name:   "map",
			Input:  Map{Simple("modules"): Array{}},
			Expect: []byte("%1\r\n+modules\r\n*0\r\n"),
			err:    nil,
		},
	}

	buf := bytes.Buffer{}
	enc := NewEncoder(&buf)
	for _, tc := range tTable {
		err := enc.Encode(tc.Input)
		if assert.Equal(t, tc.err, err) {
			actual := buf.Bytes()
			assert.Equal(t, tc.Expect, actual)
		}
		buf.Reset()
	}

}
