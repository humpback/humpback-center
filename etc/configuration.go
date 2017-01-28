package etc

import "github.com/humpback/gounits/logger"
import "gopkg.in/yaml.v2"

import (
	"io/ioutil"
	"os"
)

type Configuration struct {

	//base options
	Version string `yaml:"version"`
	PIDFile string `yaml:"pidfile"`
	//service discovery options
	Discovery struct {
		URIs      string `yaml:"uris"`
		Cluster   string `yaml:"cluster"`
		Heartbeat string `yaml:"heartbeat"`
	} `yaml:"discovery"`

	//storage options
	Storage struct {
		Mongodb struct {
			URIs string `yaml:"uris"`
		} `yaml:"mongodb,omitempty"`
	} `yaml:"storage"`

	//api options
	API struct {
		Hosts      []string `yaml:"hosts"`
		EnableCors bool     `yaml:"enablecors"`
	} `yaml:"api"`

	//log options
	Logger struct {
		LogFile  string `yaml:"logfile"`
		LogLevel string `yaml:"loglevel"`
		LogSize  int64  `yaml:"logsize"`
	} `yaml:"logger"`
}

func NewConfiguration(file string) (*Configuration, error) {

	fd, err := os.OpenFile(file, os.O_RDWR, 0777)
	if err != nil {
		return nil, err
	}

	defer fd.Close()
	data, err := ioutil.ReadAll(fd)
	if err != nil {
		return nil, err
	}

	configuration := &Configuration{}
	if err := yaml.Unmarshal([]byte(data), configuration); err != nil {
		return nil, err
	}
	return configuration, nil
}

func (configuration *Configuration) GetLogger() *logger.Args {

	return &logger.Args{
		FileName: configuration.Logger.LogFile,
		Level:    configuration.Logger.LogLevel,
		MaxSize:  configuration.Logger.LogSize,
	}
}
