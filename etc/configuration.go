package etc

import "gopkg.in/yaml.v2"

import (
	"io/ioutil"
	"os"
)

type Configuration struct {

	//基本配置参数
	Version string `yaml:"version"`
	PidFile string `yaml:"pidfile"`

	//服务发现配置
	Discovery struct {
		URIs      string `yaml:"uris"`
		SysPath   string `yaml:"syspath"`
		Heartbeat string `yaml:"heartbeat"`
	} `yaml:"discovery"`

	//持久化存储配置(目前暂支持mongo)
	Storage struct {
		Mongodb struct {
			Hosts    string            `yaml:"hosts"`
			DataBase string            `yaml:"database"`
			Auth     map[string]string `yaml:"auth,omitempty"`
		} `yaml:"mongodb,omitempty"`
	} `yaml:"storage"`

	//API服务配置
	APIServer struct {
		Bind string              `yaml:"bind"`
		Cors map[string][]string `yaml:"cors"`
	} `yaml:"apiserver"`

	//日志选项配置
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

	c := &Configuration{}
	if err := yaml.Unmarshal([]byte(data), c); err != nil {
		return nil, err
	}
	return c, nil
}
