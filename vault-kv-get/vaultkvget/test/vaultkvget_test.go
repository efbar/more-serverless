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
	"github.com/efbar/more-serverless/vault-kv-get/vaultkvget"
	vault "github.com/hashicorp/vault/api"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

const vaultServerPort = "1234"

func TestFunc(t *testing.T) {

	tt := []struct {
		loglevel    string
		contentType string
	}{
		{
			loglevel:    "INFO",
			contentType: "text/plain",
		},
		{
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
		"endpoint": "http://127.0.0.1:" + vaultServerPort,
		"token":    "root",
		"path":     "secret/data/test",
	}

	conf := vault.DefaultConfig()
	conf.Address = "http://127.0.0.1:" + vaultServerPort
	vaultClient, err := vault.NewClient(conf)
	if err != nil {
		t.Logf("Can't initiate client: %s", err.Error())
	}

	vaultClient.SetToken("root")

	mountPath := "secret-kv/"

	mountInput := &vault.MountInput{
		Type: "kv-v2",
	}

	err = vaultClient.Sys().Mount(mountPath, mountInput)
	if err != nil {
		t.Logf(fmt.Sprintf("Error in mounting path %s: %s", mountPath, err.Error()))
	}

	path := "secret/data/test"
	data := map[string]interface{}{
		"foo": "bar",
	}

	data2 := map[string]interface{}{
		"data":    data,
		"options": map[string]interface{}{},
	}

	secret, err := vaultClient.Logical().Write(path, data2)
	if err != nil {
		t.Fatalf(fmt.Sprintf("Error creating path at %s: %s", path, err))
	}
	if secret == nil {
		t.Fatalf(fmt.Sprintf("Error writing data to: %s", path))
	}
	t.Log(secret)

	for _, tr := range tt {

		jsonBody, _ := json.Marshal(jsonStruct)
		req := httptest.NewRequest("GET", "/", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", tr.contentType)

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(vaultkvget.Serve)

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
		Image: containerName,
		ExposedPorts: nat.PortSet{
			"8200/tcp": struct{}{},
		},
		Cmd: []string{"server", "-dev", fmt.Sprintf("-dev-root-token-id=%s", "root"), "-dev-listen-address=0.0.0.0:8200"},
	}

	hostConfig := &container.HostConfig{
		Binds: []string{
			"/var/run/docker.sock:/var/run/docker.sock",
		},
		PortBindings: nat.PortMap{
			"8200/tcp": []nat.PortBinding{
				{
					HostIP:   "127.0.0.1",
					HostPort: vaultServerPort,
				},
			},
		},
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

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
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
