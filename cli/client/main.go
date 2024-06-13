package main

import (
	"github.com/odit-bit/kv2/cli/client/cmd"
	"github.com/spf13/cobra"
)

func main() {
	app := cobra.Command{}
	app.AddCommand(&cmd.SetCMD, &cmd.GetCMD, &cmd.Health)
	app.Execute()
}
