package etc

import "github.com/humpback/gounits/logger"
import "gopkg.in/yaml.v2"
import "humpback-center/notify"

import (
	"io/ioutil"
	"os"
)

// Configuration is exported
type Configuration struct {

	//base options
	Version string `yaml:"version"`
	PIDFile string `yaml:"pidfile"`
	SiteAPI string `yaml:"siteapi"`

	Cluster struct {
		//driver opts
		DriverOpts []string `yaml:"opts"`
		//service discovery opts
		Discovery struct {
			URIs      string `yaml:"uris"`
			Cluster   string `yaml:"cluster"`
			Heartbeat string `yaml:"heartbeat"`
		} `yaml:"discovery"`
	} `yaml:"cluster"`

	//api options
	API struct {
		Hosts      []string `yaml:"hosts"`
		EnableCors bool     `yaml:"enablecors"`
	} `yaml:"api"`

	Notifications notify.Notifications `yaml:"notifications,omitempty"`

	//log options
	Logger struct {
		LogFile  string `yaml:"logfile"`
		LogLevel string `yaml:"loglevel"`
		LogSize  int64  `yaml:"logsize"`
	} `yaml:"logger"`
}

// NewConfiguration is exported
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

	conf := &Configuration{}
	if err := yaml.Unmarshal([]byte(data), conf); err != nil {
		return nil, err
	}

	if err := conf.ParseEnv(); err != nil {
		return nil, err
	}
	return conf, nil
}

// GetNotificationsEndPoints is exported
func (conf *Configuration) GetNotificationsEndPoints() []notify.EndPoint {

	return conf.Notifications.EndPoints
}

// GetLogger is exported
func (conf *Configuration) GetLogger() *logger.Args {

	return &logger.Args{
		FileName: conf.Logger.LogFile,
		Level:    conf.Logger.LogLevel,
		MaxSize:  conf.Logger.LogSize,
	}
}
