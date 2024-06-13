package client

import (
	"bufio"
	"io"

	pr "github.com/odit-bit/kv2/pkg/protocol"
)

// type SetItem struct {
// 	Key      []byte
// 	Value    []byte
// 	TimeUnit string
// 	Expire   int64
// }

type SetOpt struct {
	TimeUnit []byte
	Expire   int64
}

func (cli *Client) Set(key, value []byte, opts *SetOpt) (string, error) {
	conn, err := cli.getConn()
	if err != nil {
		return "", err
	}
	defer cli.putConn(conn)
	if err := SendSetCMD(conn, key, value, opts); err != nil {
		return "", err
	}

	// wait to parse response
	reader := bufio.NewReader(conn)
	_, err = pr.ReadSETCmdResponse(reader)
	if err != nil {
		if err != io.EOF {
			return "", err
		}
	}
	return "", err
}

func (cli *Client) Get(key []byte) ([]byte, error) {
	conn, err := cli.getConn()
	if err != nil {
		// return &getResponse{err: err}
		return nil, err
	}
	defer cli.putConn(conn)
	if err := SendGetCMD(conn, []byte(key)); err != nil {
		return nil, err
	}

	buf := bufio.NewReader(conn)
	return pr.ReadGetCMDResponse(buf)
}
