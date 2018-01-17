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
	Version      string `yaml:"version"  json:"version"`
	PIDFile      string `yaml:"pidfile" json:"pidfile"`
	RetryStartup bool   `yaml:"retrystartup" json:"retrystartup"`
	SiteAPI      string `yaml:"siteapi" json:"siteapi"`

	Cluster struct {
		//driver opts
		DriverOpts []string `yaml:"opts" json:"opts"`
		//service discovery opts
		Discovery struct {
			URIs      string `yaml:"uris" json:"uris"`
			Cluster   string `yaml:"cluster" json:"cluster"`
			Heartbeat string `yaml:"heartbeat" json:"heartbeat"`
		} `yaml:"discovery" json:"discovery"`
	} `yaml:"cluster" json:"cluster"`

	//api options
	API struct {
		Hosts      []string `yaml:"hosts" json:"hosts"`
		EnableCors bool     `yaml:"enablecors" json:"enablecors"`
	} `yaml:"api" json:"api"`

	Notifications notify.Notifications `yaml:"notifications,omitempty" json:"notifications,omitempty"`

	//log options
	Logger struct {
		LogFile  string `yaml:"logfile" json:"logfile"`
		LogLevel string `yaml:"loglevel" json:"loglevel"`
		LogSize  int64  `yaml:"logsize" json:"logsize"`
	} `yaml:"logger" json:"logger"`
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
