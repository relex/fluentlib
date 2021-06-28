package cmd

import (
	"bufio"
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
		Address:        "localhost:24224",
		Secret:         "guess",
		TLS:            true,
		RandomAuthFail: 0.0,
		RandomConnKill: 0.0,
		RandomNoAnswer: 0.0,
	},
}

func (cmd *serverCmdState) Run(args []string) {
	outwriter := bufio.NewWriter(os.Stdout)
	defer outwriter.Flush()

	testServer := server.LaunchServer(logger.Root(), cmd.Config, outwriter)

	sigChan := make(chan os.Signal, 10)
	signal.Notify(sigChan, syscall.SIGINT)
	signal.Notify(sigChan, syscall.SIGTERM)

	s := <-sigChan
	logger.Infof("server received %v, stopping", s)

	testServer.Shutdown()
	logger.Info("server stopped")
}
