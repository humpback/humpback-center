package etc

import "github.com/humpback/gounits/convert"

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	ERRConfigurationParseEnv = errors.New("configuration parseEnv error")
)

// ParseEnv is exported
func (conf *Configuration) ParseEnv() error {

	siteAPI := os.Getenv("HUMPBACK_SITEAPI")
	if siteAPI != "" {
		if _, err := url.Parse(siteAPI); err != nil {
			return fmt.Errorf("%s, HUMPBACK_SITEAPI %s", ERRConfigurationParseEnv.Error(), err.Error())
		}
		conf.SiteAPI = siteAPI
	}

	if err := parseClusterEnv(conf); err != nil {
		return err
	}

	if err := parseAPIEnv(conf); err != nil {
		return err
	}

	if err := parseLogEnv(conf); err != nil {
		return err
	}

	return nil
}

func parseClusterEnv(conf *Configuration) error {

	driverOpts := convert.ConvertKVStringSliceToMap(conf.Cluster.DriverOpts)
	clusterLocation := os.Getenv("CENTER_CLUSTER_LOCATION")
	if clusterLocation != "" {
		driverOpts["location"] = clusterLocation
	}

	cacheRoot := os.Getenv("CENTER_CLUSTER_CACHEROOT")
	if cacheRoot != "" {
		if _, err := filepath.Abs(cacheRoot); err != nil {
			return fmt.Errorf("%s, CENTER_CLUSTER_CACHEROOT %s", ERRConfigurationParseEnv.Error(), err.Error())
		}
		driverOpts["cacheroot"] = cacheRoot
	}

	overCommit := os.Getenv("CENTER_CLUSTER_OVERCOMMIT")
	if overCommit != "" {
		if _, err := strconv.ParseFloat(overCommit, 2); err != nil {
			return fmt.Errorf("%s, CENTER_CLUSTER_OVERCOMMIT %s", ERRConfigurationParseEnv.Error(), err.Error())
		}
		driverOpts["overcommit"] = overCommit
	}

	recoveryInterval := os.Getenv("CENTER_CLUSTER_RECOVERYINTERVAL")
	if recoveryInterval != "" {
		if _, err := time.ParseDuration(recoveryInterval); err != nil {
			return fmt.Errorf("%s, CENTER_CLUSTER_RECOVERYINTERVAL %s", ERRConfigurationParseEnv.Error(), err.Error())
		}
		driverOpts["recoveryinterval"] = recoveryInterval
	}

	createRetry := os.Getenv("CENTER_CLUSTER_CREATERETRY")
	if createRetry != "" {
		if _, err := strconv.Atoi(createRetry); err != nil {
			return fmt.Errorf("%s, CENTER_CLUSTER_CREATERETRY %s", ERRConfigurationParseEnv.Error(), err.Error())
		}
		driverOpts["createretry"] = createRetry
	}

	migrateDelay := os.Getenv("CENTER_CLUSTER_MIGRATEDELAY")
	if migrateDelay != "" {
		if _, err := time.ParseDuration(migrateDelay); err != nil {
			return fmt.Errorf("%s, CENTER_CLUSTER_MIGRATEDELAY %s", ERRConfigurationParseEnv.Error(), err.Error())
		}
		driverOpts["migratedelay"] = migrateDelay
	}
	conf.Cluster.DriverOpts = convert.ConvertMapToKVStringSlice(driverOpts)

	clusterURIs := os.Getenv("DOCKER_CLUSTER_URIS")
	if clusterURIs != "" {
		conf.Cluster.Discovery.URIs = clusterURIs
	}

	clusterName := os.Getenv("DOCKER_CLUSTER_NAME")
	if clusterName != "" {
		conf.Cluster.Discovery.Cluster = clusterName
	}

	clusterHeartBeat := os.Getenv("DOCKER_CLUSTER_HEARTBEAT")
	if clusterHeartBeat != "" {
		if _, err := time.ParseDuration(clusterHeartBeat); err != nil {
			return fmt.Errorf("%s, DOCKER_CLUSTER_HEARTBEAT %s", ERRConfigurationParseEnv.Error(), err.Error())
		}
		conf.Cluster.Discovery.Heartbeat = clusterHeartBeat
	}
	return nil
}

func parseAPIEnv(conf *Configuration) error {

	listenPort := os.Getenv("CENTER_LISTEN_PORT")
	if listenPort != "" {
		var (
			bindIPAddr string
			bindPort   string
		)
		bindArray := strings.SplitN(listenPort, ":", 2)
		if len(bindArray) == 1 {
			bindPort = bindArray[0]
		} else {
			bindIPAddr := bindArray[0]
			bindPort = bindArray[1]
			if len(bindIPAddr) > 0 {
				if _, err := net.ResolveIPAddr("tcp", bindIPAddr); err != nil {
					return fmt.Errorf("%s, CENTER_LISTEN_PORT host ipaddr error, %s", ERRConfigurationParseEnv.Error(), err.Error())
				}
			}
		}
		nPort, err := strconv.Atoi(bindPort)
		if err != nil {
			return fmt.Errorf("%s, CENTER_LISTEN_PORT %s", ERRConfigurationParseEnv.Error(), err.Error())
		}
		if nPort <= 0 || nPort > 65535 {
			return fmt.Errorf("%s, CENTER_LISTEN_PORT range invalid", ERRConfigurationParseEnv.Error())
		}
		conf.API.Hosts = []string{bindIPAddr + ":" + bindPort}
	}

	enableCore := os.Getenv("CENTER_API_ENABLECORS")
	if enableCore != "" {
		ret, err := strconv.ParseBool(enableCore)
		if err != nil {
			return fmt.Errorf("%s, CENTER_API_ENABLECORS %s", ERRConfigurationParseEnv.Error(), err.Error())
		}
		conf.API.EnableCors = ret
	}
	return nil
}

func parseLogEnv(conf *Configuration) error {

	logFile := os.Getenv("CENTER_LOG_FILE")
	if logFile != "" {
		if _, err := filepath.Abs(logFile); err != nil {
			return fmt.Errorf("%s, CENTER_LOG_FILE %s", ERRConfigurationParseEnv.Error(), err.Error())
		}
		conf.Logger.LogFile = logFile
	}

	logLevel := os.Getenv("CENTER_LOG_LEVEL")
	if logLevel != "" {
		conf.Logger.LogLevel = logLevel
	}

	logSize := os.Getenv("CENTER_LOG_SIZE")
	if logSize != "" {
		lSize, err := strconv.Atoi(logSize)
		if err != nil {
			return fmt.Errorf("%s, CENTER_LOG_SIZE %s", ERRConfigurationParseEnv.Error(), err.Error())
		}
		conf.Logger.LogSize = (int64)(lSize)
	}
	return nil
}
