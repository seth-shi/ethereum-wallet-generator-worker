package main

import (
	"os"

	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/master"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/utils"
	"github.com/seth-shi/ethereum-wallet-generator-worker/internal/worker"
	"github.com/urfave/cli/v2"
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
