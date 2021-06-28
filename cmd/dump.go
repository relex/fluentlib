package cmd

import (
	"github.com/relex/fluentlib/dump"
	"github.com/relex/gotils/logger"
)

type dumpCmdState struct {
}

var dumpCmd = dumpCmdState{}

func (cmd *dumpCmdState) Run(args []string) {
	if len(args) < 1 {
		logger.Fatal("requires at least one file or directory")
	}
	dump.PrintFileOrDirectories(args)
}
