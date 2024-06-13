package cmd

import (
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/odit-bit/kv2/server"
	"github.com/odit-bit/kv2/store/x"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// COMMAND

// func init() {
// 	var (
// 		port      string
// 		host      string
// 		debug     bool
// 		cacheSize int
// 	)

// 	Run.PersistentFlags().StringVarP(&port, "port", "p", "6969", "port bind")
// 	// viper.BindPFlag("port", Run.Flags().Lookup("port"))
// 	viper.BindEnv("port", "PORT")
// 	// viper.SetDefault("port", "6969")

// 	Run.PersistentFlags().StringVar(&host, "host", "", "host name")
// 	// viper.BindPFlag("host", Run.Flags().Lookup("host"))
// 	viper.BindEnv("host", "HOST")
// 	// viper.SetDefault("host", "")

// 	Run.PersistentFlags().BoolVarP(&debug, "debug", "", false, "debug mode")
// 	// viper.BindPFlag("debug", Run.Flags().Lookup("debug"))
// 	viper.BindEnv("debug", "DEBUG_MODE")
// 	// viper.SetDefault("debug", false)

// 	Run.PersistentFlags().IntVarP(&cacheSize, "size", "s", 512*humanize.MByte, "cache capacity in byte")
// 	// viper.BindPFlag("size", Run.Flags().Lookup("size"))
// 	viper.BindEnv("size", "CACHE_SIZE")
// 	// viper.SetDefault("size", 1*humanize.GByte)
// }

var Run = cobra.Command{
	Use:     "run",
	Example: "",
	Version: "1",
	Args:    cobra.MaximumNArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		viper.BindPFlag("port", cmd.Flags().Lookup("port"))
		viper.BindPFlag("host", cmd.Flags().Lookup("host"))
		viper.BindPFlag("debug", cmd.Flags().Lookup("debug"))
		viper.BindPFlag("size", cmd.Flags().Lookup("size"))

		host := viper.GetString("host")
		port := viper.GetString("port")
		debug := viper.GetBool("debug")
		size := viper.GetInt("size")

		runServer(host, port, debug, size)
	},
}

func runServer(host, port string, debug bool, size int) {

	addr := host + ":" + port
	db := x.NewFastcache(size)
	logger := logger(debug)

	srv := server.New(addr, db, logger)
	var wg sync.WaitGroup
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	wg.Add(1)
	go func() {
		v, ok := <-sig
		if ok {
			logger.Info("got signal", "type", v)
		}

		//close server
		if err := srv.ShutDown(); err != nil {
			logger.Error(err.Error())
		}

		//closing db
		_ = db.Close()
		wg.Done()
	}()

	logger.Info("server start", "addr", srv.Address())
	err := srv.Serve()
	wg.Wait()

	if err != nil {
		logger.Error("server exit", "err", err)
		os.Exit(2)
	}

}

func logger(debug bool) *slog.Logger {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: debug,
		Level:     level,
	}))
	return logger
}
