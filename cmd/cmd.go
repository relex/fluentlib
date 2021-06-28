// Package cmd provides list of commands for tools
package cmd

import (
	"github.com/relex/gotils/config"
)

func init() {
	config.AddParentCmdWithArgs("", "Tools for Fluentd / Fluent Bit", nil, nil, nil)
	config.AddCmdWithArgs("dump <path-to-files-or-dirs>...", "Dump given files or dirs. Support Fluent Bit chunk files (.flb) and Fluentd Forward messages in msgpack format", &dumpCmd, dumpCmd.Run)
	config.AddCmdWithArgs("server <output_file>", "Run a test server for Fluentd Forward Protocol and output logs in JSON.", &serverCmd, serverCmd.Run)
}

// Execute parses command-line and executes the root command
func Execute() {
	// trigger init

	config.Execute()
}
