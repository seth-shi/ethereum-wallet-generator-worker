package main

import (
	"github.com/seth-shi/ethereum-wallet-generator-nodes/internal"
	"github.com/urfave/cli/v2"
	"os"
)

var (
	Master *internal.Master
	Node   *internal.Node
)

func main() {

	app := &cli.App{
		Commands: []*cli.Command{
			masterCommand,
			nodeCommand,
		},
	}

	internal.MustError(app.Run(os.Args))
}
