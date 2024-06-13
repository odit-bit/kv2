package main

import (
	"github.com/dustin/go-humanize"
	"github.com/odit-bit/kv2/cli/server/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	var (
		port      string
		host      string
		debug     bool
		cacheSize int
	)

	app.PersistentFlags().StringVarP(&port, "port", "p", "6969", "port bind")
	// viper.BindPFlag("port", Run.Flags().Lookup("port"))
	viper.BindEnv("port", "PORT")
	// viper.SetDefault("port", "6969")

	app.PersistentFlags().StringVar(&host, "host", "", "host name")
	viper.BindEnv("host", "HOST")

	app.PersistentFlags().BoolVarP(&debug, "debug", "", false, "debug mode")
	viper.BindEnv("debug", "DEBUG_MODE")

	app.PersistentFlags().IntVarP(&cacheSize, "size", "s", 512*humanize.MByte, "cache capacity in byte")
	viper.BindEnv("size", "CACHE_SIZE")
}

var app = cobra.Command{}

func main() {
	app.AddCommand(&cmd.Run, &cmd.Health)
	app.Execute()
}
