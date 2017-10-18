package main

import "github.com/humpback/gounits/system"
import "humpback-center/server"

import (
	"log"
	"os"
)

func main() {

	service, err := server.NewCenterService()
	if err != nil {
		log.Printf("service error:%s\n", err.Error())
		os.Exit(system.ErrorExitCode(err))
	}

	defer func() {
		service.Stop()
		os.Exit(0)
	}()

	if err := service.Startup(); err != nil {
		log.Printf("service start error:%s\n", err.Error())
		os.Exit(system.ErrorExitCode(err))
	}
	system.InitSignal(nil)
}
