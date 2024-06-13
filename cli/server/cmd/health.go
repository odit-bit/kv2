package cmd

import (
	"log"
	"net"
	"os"
	"time"

	"github.com/odit-bit/kv2/pkg/protocol"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Health = cobra.Command{
	Use: "health",
	Run: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("port", cmd.Flags().Lookup("port"))
		viper.BindPFlag("host", cmd.Flags().Lookup("host"))

		host := viper.GetString("host")
		port := viper.GetString("port")

		addr := host + ":" + port
		if err := health(addr); err != nil {
			log.Println(err)
			os.Exit(2)
		}
	},
}

func health(addr string) error {
	//Connect
	var conn net.Conn
	var err error
	count := 3
	for count != 0 {
		conn, err = net.Dial("tcp", addr)
		if err != nil {
			log.Printf("connnecting %s, retrying...", err)
			count--
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		return err
	}
	defer conn.Close()

	//PING
	count = 3
	for count != 0 {
		conn.SetDeadline(time.Now().Add(5 * time.Second))
		if err := protocol.Write(conn, protocol.Array{
			protocol.Bulk("HEALTH"),
			protocol.Bulk("PING"),
		}); err != nil {
			log.Printf("check server health %s, %s retrying...", conn.RemoteAddr().String(), err)
			count--
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}

	log.Println("server health")
	return nil
}
