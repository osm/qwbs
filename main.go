package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/osm/qwbs/internal/config"
	"github.com/osm/qwbs/internal/server"
	"github.com/osm/qwbs/internal/version"
)

func main() {
	configFile := flag.String("config-file", "./qwbs.conf", "Path to config file")
	flag.Parse()

	conf, err := config.FromFile(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse config file: %v\n", err)
		os.Exit(1)
	}

	var logLevel slog.LevelVar
	logLevel.Set(slog.LevelInfo)
	if conf.Debug {
		logLevel.Set(slog.LevelDebug)
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: &logLevel}))

	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	logger.Info(version.Name(),
		"listen-addr", conf.ListenAddress,
		"version", version.Short(),
		"writers", len(conf.Writers))

	srv := server.New(logger, conf.ListenAddress, conf.MasterAddresses, conf.Writers)
	if err := srv.ListenAndServe(ctx); err != nil {
		logger.Error("ListenAndServe failed", "error", err)
		os.Exit(1)
	}
}
