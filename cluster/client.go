package cluster

import "common/models"
import "github.com/humpback/gounits/httpx"
import "github.com/docker/docker/api/types"
import ctypes "humpback-center/cluster/types"

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

// Client is exported
type Client struct {
	apiAddr string
	c       *httpx.HttpClient
}

// NewClient is exported
func NewClient(apiAddr string) *Client {

	client := httpx.NewClient().
		SetTransport(&http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   45 * time.Second,
				KeepAlive: 90 * time.Second,
			}).DialContext,
			DisableKeepAlives:     false,
			MaxIdleConns:          10,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   http.DefaultTransport.(*http.Transport).TLSHandshakeTimeout,
			ExpectContinueTimeout: http.DefaultTransport.(*http.Transport).ExpectContinueTimeout,
		})

	return &Client{
		apiAddr: apiAddr,
		c:       client,
	}
}

// Close is exported
// client close
func (client *Client) Close() {

	client.c.Close()
}

// GetDockerInfoRequest is exported
// get docker node info
func (client *Client) GetDockerInfoRequest(ctx context.Context) (*types.Info, error) {

	respSpecs, err := client.c.Get(ctx, "http://"+client.apiAddr+"/v1/dockerinfo", nil, nil)
	if err != nil {
		return nil, err
	}

	defer respSpecs.Close()
	if respSpecs.StatusCode() >= http.StatusBadRequest {
		return nil, fmt.Errorf("dockerinfo request, %s", ctypes.ParseHTTPResponseError(respSpecs))
	}

	dockerInfo := &types.Info{}
	if err := respSpecs.JSON(dockerInfo); err != nil {
		return nil, err
	}
	return dockerInfo, nil
}

// GetContainerRequest is exported
// get a container type info.
func (client *Client) GetContainerRequest(ctx context.Context, containerid string) (*types.ContainerJSON, error) {

	query := map[string][]string{"originaldata": []string{"true"}}
	respContainer, err := client.c.Get(ctx, "http://"+client.apiAddr+"/v1/containers/"+containerid, query, nil)
	if err != nil {
		return nil, err
	}

	defer respContainer.Close()
	if respContainer.StatusCode() >= http.StatusBadRequest {
		return nil, fmt.Errorf("container %s request, %s", ShortContainerID(containerid), ctypes.ParseHTTPResponseError(respContainer))
	}

	containerJSON := &types.ContainerJSON{}
	if err := respContainer.JSON(containerJSON); err != nil {
		return nil, err
	}
	return containerJSON, nil
}

// GetContainersRequest is exported
// return all containers info.
func (client *Client) GetContainersRequest(ctx context.Context) ([]types.Container, error) {

	query := map[string][]string{"all": []string{"true"}}
	respContainers, err := client.c.Get(ctx, "http://"+client.apiAddr+"/v1/containers", query, nil)
	if err != nil {
		return nil, err
	}

	defer respContainers.Close()
	if respContainers.StatusCode() >= http.StatusBadRequest {
		return nil, fmt.Errorf("containers request, %s", ctypes.ParseHTTPResponseError(respContainers))
	}

	allContainers := []types.Container{}
	if err := respContainers.JSON(&allContainers); err != nil {
		return nil, err
	}
	return allContainers, nil
}

// CreateContainerRequest is exported
// create a container request.
func (client *Client) CreateContainerRequest(ctx context.Context, config models.Container) (*ctypes.CreateContainerResponse, error) {

	respCreated, err := client.c.PostJSON(ctx, "http://"+client.apiAddr+"/v1/containers", nil, config, nil)
	if err != nil {
		return nil, err
	}

	defer respCreated.Close()
	if respCreated.StatusCode() >= http.StatusBadRequest {
		return nil, fmt.Errorf("create container %s request, %s", config.Name, ctypes.ParseHTTPResponseError(respCreated))
	}

	createContainerResponse := &ctypes.CreateContainerResponse{}
	if err := respCreated.JSON(createContainerResponse); err != nil {
		return nil, err
	}
	return createContainerResponse, nil
}

// RemoveContainerRequest is exported
// remove a container request.
func (client *Client) RemoveContainerRequest(ctx context.Context, containerid string) error {

	query := map[string][]string{"force": []string{"true"}}
	respRemoved, err := client.c.Delete(ctx, "http://"+client.apiAddr+"/v1/containers/"+containerid, query, nil)
	if err != nil {
		return err
	}

	defer respRemoved.Close()
	if respRemoved.StatusCode() >= http.StatusBadRequest {
		return fmt.Errorf("remove container %s request, %s", ShortContainerID(containerid), ctypes.ParseHTTPResponseError(respRemoved))
	}
	return nil
}

// OperateContainerRequest is exported
// operate a container request.
func (client *Client) OperateContainerRequest(ctx context.Context, operate models.ContainerOperate) error {

	respOperated, err := client.c.PutJSON(ctx, "http://"+client.apiAddr+"/v1/containers", nil, operate, nil)
	if err != nil {
		return err
	}

	defer respOperated.Close()
	if respOperated.StatusCode() >= http.StatusBadRequest {
		return fmt.Errorf("%s container %s request, %s", operate.Action, ShortContainerID(operate.Container), ctypes.ParseHTTPResponseError(respOperated))
	}
	return nil
}

// UpgradeContainerRequest is exported
// upgrade a container request.
func (client *Client) UpgradeContainerRequest(ctx context.Context, operate models.ContainerOperate) (*ctypes.UpgradeContainerResponse, error) {

	respUpgraded, err := client.c.PutJSON(ctx, "http://"+client.apiAddr+"/v1/containers", nil, operate, nil)
	if err != nil {
		return nil, err
	}

	defer respUpgraded.Close()
	if respUpgraded.StatusCode() >= http.StatusBadRequest {
		return nil, fmt.Errorf("upgrate container %s request, %s", ShortContainerID(operate.Container), ctypes.ParseHTTPResponseError(respUpgraded))
	}

	upgradeContainerResponse := &ctypes.UpgradeContainerResponse{}
	if err := respUpgraded.JSON(upgradeContainerResponse); err != nil {
		return nil, err
	}
	return upgradeContainerResponse, nil
}
