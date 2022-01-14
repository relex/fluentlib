package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/relex/fluentlib/server"
	"github.com/relex/fluentlib/server/receivers"
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
		SplitOutputKeys:   []string{"app", "level", "pnum"},
		SplitOutputPath:   "",
		SplitStrictMode:   false,
		RandomNoHandshake: 0.0,
		RandomFailAuth:    0.0,
		RandomNoReceiving: 0.0,
		RandomNoResponse:  0.0,
		RandomKillConn:    0.0,
	},
}

func (cmd *serverCmdState) Run(args []string) {
	var receiver receivers.Receiver
	if len(serverCmd.SplitOutputPath) > 0 {
		if err := receivers.VerifySplittingFilePath(serverCmd.SplitOutputPath); err != nil {
			logger.Fatal("invalid split_output_path: ", err.Error())
		}

		logger.WithFields(logger.Fields{
			"keys":   serverCmd.SplitOutputKeys,
			"path":   serverCmd.SplitOutputPath,
			"strict": serverCmd.SplitStrictMode,
		}).Infof("use split output")
		receiver = receivers.NewSplittingFileWriter(serverCmd.SplitOutputKeys, serverCmd.SplitOutputPath, serverCmd.SplitStrictMode)
	} else {
		logger.Infof("use message output")
		receiver = receivers.NewMessageWriter(os.Stdout)
	}
	srv, _ := server.LaunchServer(logger.Root(), cmd.Config, receiver)

	sigChan := make(chan os.Signal, 10)
	signal.Notify(sigChan, syscall.SIGINT)
	signal.Notify(sigChan, syscall.SIGTERM)

	s := <-sigChan
	logger.Infof("server received %v, stopping", s)

	srv.Shutdown()
	logger.Info("server stopped")
	logger.Exit(0)
}
