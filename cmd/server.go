package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/relex/fluentlib/server"
	"github.com/relex/gotils/logger"
)

type serverCmdState struct {
	server.Config
}

var serverCmd = serverCmdState{
	Config: server.Config{
		Address:           "localhost:24224",
		Secret:            "guess",
		TLS:               true,
		RandomNoHandshake: 0.0,
		RandomFailAuth:    0.0,
		RandomNoReceiving: 0.0,
		RandomNoResponse:  0.0,
		RandomKillConn:    0.0,
	},
}

func (cmd *serverCmdState) Run(args []string) {
	srv, _ := server.LaunchServer(logger.Root(), cmd.Config, server.NewMessageWriter(os.Stdout))

	sigChan := make(chan os.Signal, 10)
	signal.Notify(sigChan, syscall.SIGINT)
	signal.Notify(sigChan, syscall.SIGTERM)

	s := <-sigChan
	logger.Infof("server received %v, stopping", s)

	srv.Shutdown()
	logger.Info("server stopped")
}
