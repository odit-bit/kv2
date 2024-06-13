package cmd

import (
	"github.com/odit-bit/kv2/server"
	"github.com/spf13/cobra"
)

var EdgeCMD = cobra.Command{
	Use:     "edge",
	Short:   "run edge server , not stable",
	Args:    cobra.ExactArgs(0),
	Version: "EDGE",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := server.NewEdge()
		return app.Run()
	},
}
