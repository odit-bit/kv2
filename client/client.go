package client

import (
	"crypto/tls"
	"log"
	"net"
)

type Client struct {
	cp *connPool
}

type Config struct {
	MinConn int
	MaxConn int
}

func New(addr string) *Client {

	dial := func() (net.Conn, error) {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			return nil, err
		}
		return conn, nil
	}
	conf := Config{
		MinConn: 5,
		MaxConn: 20,
	}

	cp, err := NewConnPool(conf.MinConn, conf.MaxConn, dial)
	if err != nil {
		log.Fatal(err)
	}

	cli := Client{
		cp: cp,
	}

	return &cli
}

func NewTLS(addr string, conf *tls.Config) *Client {
	dial := func() (net.Conn, error) {
		return tls.Dial("tcp", addr, conf)
	}

	cp, err := NewConnPool(5, 10, dial)
	if err != nil {
		log.Fatal(err)
	}

	cli := Client{
		cp: cp,
	}

	return &cli
}

func (cli *Client) getConn() (net.Conn, error) {
	// get connection
	max := 3
	var err error
	var conn net.Conn

	for {
		conn, err = cli.cp.Get()
		if err != nil {
			if max == 0 {
				return nil, err
			}
			max--
			continue
		}
		return conn, nil
	}
}

func (cli *Client) putConn(conn net.Conn) {
	cli.cp.Put(conn)
}

func (cli *Client) Close() error {
	cli.cp.Close()
	return nil
}
