package cmd

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/odit-bit/kv2/client"
	"github.com/odit-bit/kv2/pkg/protocol"
	"github.com/spf13/cobra"
)

var GetCMD = cobra.Command{
	Use:     "GET key",
	Example: "GET foo",
	Args:    cobra.MinimumNArgs(1),
	Version: "0.1",

	Run: func(cmd *cobra.Command, args []string) {
		conn, err := net.Dial("tcp", "localhost:6969")
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()
		if err := protocol.Write(conn, protocol.Array{
			protocol.Bulk("GET"),
			protocol.Bulk(args[0]),
		}); err != nil {
			return
		}

		r := bufio.NewReader(conn)
		if msg, err := protocol.ReadGetCMDResponse(r); err != nil {
			fmt.Println(err)
			return
		} else {
			fmt.Println(string(msg))
			return
		}

	},
}

var SetCMD = cobra.Command{
	Use:     "SET key value",
	Example: "SET foo bar",
	Args:    cobra.MinimumNArgs(2),
	Version: "0.1",

	Run: func(cmd *cobra.Command, args []string) {
		conn, err := net.Dial("tcp", ":6969")
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()
		// dec := protocol.NewDecoder(conn)

		//build command
		sc := setCMD{
			Name:  []byte("SET"),
			Key:   []byte(args[0]),
			Value: []byte(args[1]),
		}
		if len(args) > 2 {
			sc.Unit = []byte(args[2])
			expire, err := strconv.Atoi(args[3])
			if err != nil {
				fmt.Println("expire should integer")
				return
			}
			sc.expire = int64(expire)
		}

		client.SendSetCMD(conn, sc.Key, sc.Value, &client.SetOpt{
			TimeUnit: sc.Unit,
			Expire:   sc.expire,
		})

		r := bufio.NewReader(conn)
		if msg, err := protocol.ReadSETCmdResponse(r); err != nil {
			fmt.Println(err)
			return
		} else {
			fmt.Println(msg)
			return
		}
	},
}

type setCMD struct {
	Name   []byte
	Key    []byte
	Value  []byte
	Unit   []byte
	expire int64
}
