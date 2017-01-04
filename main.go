package main

import "github.com/humpback/gounits/system"
import "github.com/humpback/humpback-center/server"

import (
	"os"
)

func main() {

	center, err := server.NewServerCenter()
	if err != nil {
		panic(err)
		os.Exit(EXITCODE_INITFAILED)
	}

	defer func() {
		center.Stop()
		os.Exit(EXITCODE_EXITED)
	}()

	if err := center.Startup(); err != nil {
		panic(err)
		os.Exit(EXITCODE_STARTFAILED)
	}
	system.InitSignal(nil)
}
