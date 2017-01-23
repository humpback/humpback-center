package server

import "github.com/humpback/humpback-center/etc"
import "github.com/humpback/humpback-center/api"
import "github.com/humpback/humpback-center/cluster"
import "github.com/humpback/humpback-center/repository"
import "github.com/humpback/humpback-center/storage"
import "github.com/humpback/gounits/logger"

import (
	"flag"
)

/*
ServerCenter
humpback center service
*/
type CenterService struct {
	APIServer       *api.Server
	Cluster         *cluster.Cluster
	RepositoryCache *repository.RepositoryCache
	DataStorage     *storage.DataStorage
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

	largs := configuration.GetLogger()
	logger.OPEN(largs)
	mongo := configuration.Storage.Mongodb
	datastorage, err := storage.NewDataStorage(mongo.URIs)
	if err != nil {
		return nil, err
	}

	cluster, err := createCluster(configuration)
	if err != nil {
		return nil, err
	}

	repositorycache := repository.NewRepositoryCache()
	router := api.NewRouter(cluster, repositorycache, configuration.API.EnableCors)
	apiserver := api.NewServer(configuration.API.Hosts, nil, router)
	return &CenterService{
		APIServer:   apiserver,
		Cluster:     cluster,
		DataStorage: datastorage,
	}, nil
}

func (service *CenterService) Startup() error {

	logger.INFO("[#service#] service start...")
	go func() {
		if err := service.APIServer.Startup(); err != nil {
			logger.ERROR("[#service#] service API start error:%s", err.Error())
		}
	}()

	groups := service.DataStorage.GetGroups()
	for _, group := range groups {
		service.Cluster.CreateGroup(group.ID, group.Servers)
	}
	return service.Cluster.Start()
}

func (service *CenterService) Stop() error {

	service.Cluster.Stop()
	service.DataStorage.Close()
	logger.INFO("[#service#] service closed.")
	logger.CLOSE()
	return nil
}
