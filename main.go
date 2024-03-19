package main

import (
	"github.com/seth-shi/ethereum-wallet-generator-nodes/internal/master"
	"github.com/seth-shi/ethereum-wallet-generator-nodes/internal/utils"
	"github.com/seth-shi/ethereum-wallet-generator-nodes/internal/worker"
	"github.com/urfave/cli/v2"
	"os"
)

var (
	Master *master.Master
	Worker *worker.Worker
)

func main() {

	app := &cli.App{
		Commands: []*cli.Command{
			masterCommand,
			workerCommand,
			decryptCommand,
		},
	}

	utils.MustError(app.Run(os.Args))
}
