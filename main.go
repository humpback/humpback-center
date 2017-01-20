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
		os.Exit(EXITCODE_INITFAILED)
	}

	defer func() {
		service.Stop()
		os.Exit(EXITCODE_EXITED)
	}()

	if err := service.Startup(); err != nil {
		panic(err)
		os.Exit(EXITCODE_STARTFAILED)
	}
	system.InitSignal(nil)
}
