package testing

import (
	"bytes"
	"context"
	"encoding/base64"
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
	"github.com/efbar/more-serverless/vault-transit/vaulttransit"
	vault "github.com/hashicorp/vault/api"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

const vaultServerPort = "1234"

func TestFunc(t *testing.T) {

	tt := []struct {
		contentType string
		action      string
	}{
		{
			contentType: "text/plain",
			action:      "encrypt",
		},
		{
			contentType: "application/json",
			action:      "decrypt",
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

	conf := vault.DefaultConfig()
	conf.Address = "http://127.0.0.1:" + vaultServerPort
	vaultClient, err := vault.NewClient(conf)
	if err != nil {
		t.Logf("Can't initiate client: %s", err.Error())
	}

	vaultClient.SetToken("root")

	mountPath := "transit/"

	mountInput := &vault.MountInput{
		Type: "transit",
	}

	err = vaultClient.Sys().Mount(mountPath, mountInput)
	if err != nil {
		t.Logf(fmt.Sprintf("Error in mounting path %s: %s", mountPath, err.Error()))
	}

	transitKey := "transit/keys/testkey"

	transitData := map[string]interface{}{
		"data":    "",
		"options": map[string]interface{}{},
	}

	secret, err := vaultClient.Logical().Write(transitKey, transitData)
	if err != nil {
		t.Fatalf(fmt.Sprintf("Error creating path at %s: %s", transitKey, err))
	}
	if secret == nil {
		t.Log(fmt.Sprintf("Success! Data written to: %s", transitKey))
	}

	for _, tr := range tt {
		t.Log(tr.action)

		data := map[string]interface{}{
			"plaintext": base64.StdEncoding.EncodeToString([]byte("foobar")),
		}
		if tr.action == "decrypt" {

			testKeyPath := "transit/encrypt/testkey"

			secret, err = vaultClient.Logical().Write(testKeyPath, data)
			if err != nil {
				t.Fatalf(fmt.Sprintf("Error creating path at %s: %s", testKeyPath, err))
			}
			if secret == nil {
				t.Log(fmt.Sprintf("Success! Data written to: %s", testKeyPath))
			}
			cypherData := secret.Data["ciphertext"]
			data = map[string]interface{}{
				"ciphertext": cypherData,
			}

		}

		testKeyPath := "transit/" + tr.action + "/testkey"

		jsonStruct := map[string]interface{}{
			"endpoint": "http://127.0.0.1:" + vaultServerPort,
			"token":    "root",
			"path":     testKeyPath,
			"data":     data,
		}

		jsonBody, _ := json.Marshal(jsonStruct)
		req := httptest.NewRequest("GET", "/", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", tr.contentType)

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(vaulttransit.Serve)

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
