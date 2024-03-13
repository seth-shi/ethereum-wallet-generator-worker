package internal

import (
	"context"
	"os/signal"
	"syscall"
)

func NewSignal() (context.Context, func()) {
	return signal.NotifyContext(context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)
}
