package server

import "github.com/humpback/humpback-center/api"
import "github.com/humpback/humpback-center/ctrl"
import "github.com/humpback/humpback-center/etc"
import "github.com/humpback/gounits/fprocess"
import "github.com/humpback/gounits/logger"

import (
	"flag"
)

/*
ServerCenter
humpback center service
*/
type CenterService struct {
	PIDFile    *fprocess.PIDFile
	APIServer  *api.Server
	Controller *ctrl.Controller
}

// NewCenterService exported
func NewCenterService() (*CenterService, error) {

	var conf string
	flag.StringVar(&conf, "f", "etc/config.yaml", "humpback center configuration file.")
	flag.Parse()
	configuration, err := etc.NewConfiguration(conf)
	if err != nil {
		return nil, err
	}

	pidfile, err := fprocess.New(configuration.PIDFile)
	if err != nil {
		return nil, err
	}

	largs := configuration.GetLogger()
	logger.OPEN(largs)
	controller, err := ctrl.NewController(configuration)
	if err != nil {
		return nil, err
	}

	apiserver := api.NewServer(configuration.API.Hosts, nil, controller, configuration.API.EnableCors)
	return &CenterService{
		PIDFile:    pidfile,
		APIServer:  apiserver,
		Controller: controller,
	}, nil
}

func (service *CenterService) Startup() error {

	logger.INFO("[#service#] service start...")
	if err := service.Controller.Initialize(); err != nil {
		return err
	}
	logger.INFO("[#service#] process %d", service.PIDFile.PID)
	//apiserver start.
	go func() {
		if err := service.APIServer.Startup(); err != nil {
			logger.ERROR("[#service#] service API start error:%s", err.Error())
		}
	}()
	return nil
}

func (service *CenterService) Stop() error {

	service.Controller.UnInitialize()
	service.PIDFile.Remove()
	logger.INFO("[#service#] service closed.")
	logger.CLOSE()
	return nil
}
