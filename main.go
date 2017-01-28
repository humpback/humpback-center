package main

import "github.com/humpback/gounits/system"
import "github.com/humpback/humpback-center/server"

import (
	"os"
)

func main() {

	service, err := server.NewCenterService()
	if err != nil {
		panic(err)
		os.Exit(system.PorcessExitCode(err))
	}

	defer func() {
		service.Stop()
		os.Exit(0)
	}()

	if err := service.Startup(); err != nil {
		panic(err)
		os.Exit(system.PorcessExitCode(err))
	}
	system.InitSignal(nil)
}
