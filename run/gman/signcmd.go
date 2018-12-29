package main

import (
	"gopkg.in/urfave/cli.v1"

	"github.com/matrix/go-matrix/run/utils"
)

var (
	signatureCommand = cli.Command{
		Action:   utils.MigrateFlags(signature),
		Name:     "sign",
		Usage:    "Start an interactive JavaScript environment",
		Flags:    nodeFlags,
		Category: "SIGNATURE COMMANDS",
		Description: `
The Gman console is an interactive shell for the JavaScript runtime environment
which exposes a node admin interface as well as the Ðapp JavaScript API.
See https://github.com/matrix/go-matrix/wiki/JavaScript-Console.`,
	}
)

// localConsole starts a new gman node, attaching a JavaScript console to it at the
// same time.
func signature(ctx *cli.Context) error {
	// Create and start the node based on the CLI flags
	node := makeFullNode(ctx)
	node.StartSign()
	return nil
}
