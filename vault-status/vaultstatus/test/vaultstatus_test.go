package testing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	network "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/efbar/more-serverless/vault-status/vaultstatus"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func TestFunc(t *testing.T) {

	tt := []struct {
		aclenabled  bool
		loglevel    string
		contentType string
	}{
		{
			aclenabled:  true,
			loglevel:    "INFO",
			contentType: "text/plain",
		},
		{
			aclenabled:  true,
			loglevel:    "INFO",
			contentType: "application/json",
		},
	}

	cli, err := client.NewClientWithOpts()
	if err != nil {
		fmt.Printf("Can't instantiate docker client: %s", err.Error())
	}

	resp, err := startContainerService(cli, "vault")
	if err != nil {
		t.Logf("Can't create container: %s", err.Error())
	}
	t.Log(resp.ID)
	defer removeContainerService(cli, resp.ID)

	jsonStruct := map[string]interface{}{
		"endpoint": "http://127.0.0.1:8200",
	}

	for _, tr := range tt {

		jsonBody, _ := json.Marshal(jsonStruct)
		req := httptest.NewRequest("GET", "/", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", tr.contentType)

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(vaultstatus.Serve)

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		fmt.Printf("response body: \n%s\n", rr.Body)
	}

}

func startContainerService(cli *client.Client, containerName string) (*container.ContainerCreateCreatedBody, error) {
	// context
	ctx := context.Background()

	// config
	config := &container.Config{
		Image:        containerName,
		ExposedPorts: nat.PortSet{"8200": struct{}{}},
		Cmd:          []string{"server", "-dev", fmt.Sprintf("-dev-root-token-id=%s", "root"), "-dev-listen-address=0.0.0.0:8200"},
	}

	hostConfig := &container.HostConfig{
		PortBindings: map[nat.Port][]nat.PortBinding{nat.Port("8200"): {{HostIP: "127.0.0.1", HostPort: "8200"}}},
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		LogConfig: container.LogConfig{
			Type:   "json-file",
			Config: map[string]string{},
		},
	}

	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}
	gatewayConfig := &network.EndpointSettings{
		Gateway: "gatewayname",
	}
	networkConfig.EndpointsConfig["bridge"] = gatewayConfig

	resp, err := cli.ContainerCreate(
		ctx,
		config,
		hostConfig,
		networkConfig,
		&v1.Platform{
			Architecture: "amd64",
			OS:           "linux",
		},
		containerName,
	)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func removeContainerService(cli *client.Client, containerName string) {
	ctx := context.Background()

	err := cli.ContainerStop(ctx, containerName, nil)
	if err != nil {
		fmt.Printf("Can't stop container: %s", err.Error())
		return
	}

	removeOptions := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}
	err = cli.ContainerRemove(ctx, containerName, removeOptions)
	if err != nil {
		fmt.Printf("Can't remove container: %s", err.Error())
		return
	}

}
