package main

import (
	"github.com/odit-bit/kv2/cli/client/cmd"
	"github.com/spf13/cobra"
)

func main() {
	var (
		port string
		host string
	)

	app := cobra.Command{}

	app.PersistentFlags().StringVarP(&port, "port", "p", "6969", "port bind")
	app.PersistentFlags().StringVar(&host, "host", "", "host name")

	app.AddCommand(&cmd.SetCMD, &cmd.GetCMD, &cmd.Health)
	app.Execute()
}
