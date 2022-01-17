package main

import (
	"github.com/relex/fluentlib/cmd"
	"github.com/relex/fluentlib/util"
	"github.com/relex/gotils/logger"
)

var version string

func main() {
	util.SeedRand() // seed rand properly for all rand.* calls

	logger.Infof("version: %s", version)

	cmd.Execute()

	logger.Exit(0)
}
