package client

import (
	"io"

	pr "github.com/odit-bit/kv2/pkg/protocol"
)

func SendSetCMD(conn io.Writer, key, value []byte, opts *SetOpt) error {
	var cmdName = []byte("SET")
	enc := pr.NewEncoder(conn)

	if len(opts.TimeUnit) != 0 {
		if err := enc.Encode(pr.Array{pr.Bulk(cmdName), pr.Bulk(key), pr.Bulk(value), pr.Bulk(opts.TimeUnit), pr.Integer(opts.Expire)}); err != nil {
			return err
		}
	} else {
		if err := enc.Encode(pr.Array{pr.Bulk(cmdName), pr.Bulk(key), pr.Bulk(value)}); err != nil {
			return err
		}
	}

	return nil
}

func SendGetCMD(w io.Writer, key []byte) error {
	return pr.Write(w, pr.Array{pr.Bulk("GET"), pr.Bulk(key)})

}
