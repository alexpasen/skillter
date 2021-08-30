package main

import (
	"os"
	"os/signal"
	"skillter/app"
	"syscall"
)

func main() {
	ok, a := app.NewApp()
	if ok != nil {
		app.Logger.Error().Msg(ok.String())
	}

	s := make(chan os.Signal, 1)
	signal.Notify(s,
		syscall.SIGABRT,
		syscall.SIGINT,
		syscall.SIGKILL,
		syscall.SIGQUIT,
		syscall.SIGTERM,
		syscall.SIGTRAP,
		syscall.SIGTSTP)

	go func() {
		for  {
			select {
			case <-s:
				a.Stop()
			}
		}
	}()


	if ok := a.Start(); ok != nil {
		app.Logger.Error().Msg(ok.String())
	}
}
