package testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	nomadnodestatus "github.com/efbar/more-serverless/nomad-node-status/nomadnodestatus"
	nomad "github.com/hashicorp/nomad/api"
	testutil "github.com/hashicorp/nomad/testutil"
)

func TestFunc(t *testing.T) {

	// Create server
	ACLEnabled := false
	Loglevel := "DEBUG"
	// ContentType := "application/json"
	ContentType := "text/plain"

	server := testutil.NewTestServer(t, func(c *testutil.TestServerConfig) {
		c.ACL.Enabled = ACLEnabled
		c.LogLevel = Loglevel
		c.DevMode = true
		// c.Ports = &testutil.PortsConfig{
		// 	HTTP: c.Ports.HTTP + 1,
		// 	RPC:  c.Ports.RPC + 1,
		// 	Serf: c.Ports.Serf + 1,
		// }
	})
	defer server.Stop()

	// Create client
	conf := nomad.DefaultConfig()
	newAddr, _ := strconv.Atoi(server.HTTPAddr)
	conf.Address = "http://" + strconv.Itoa(newAddr+1)

	client, err := nomad.NewClient(conf)
	fmt.Println(conf)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	jsonStruct := map[string]interface{}{
		"endpoint": "http://" + server.HTTPAddr,
	}

	// Get root token if ACL is enabled
	if ACLEnabled {
		root, _, err := client.ACLTokens().Bootstrap(nil)
		if err != nil {
			t.Fatalf("failed to bootstrap ACLs: %v", err)
		}
		client.SetSecretID(root.SecretID)
		t.Log("root token:", root.SecretID)

		// Register a policy
		ap := client.ACLPolicies()
		conf.SecretID = root.SecretID
		policy := &nomad.ACLPolicy{
			Name:        "test",
			Description: "test",
			Rules: `namespace "default" {
				policy = "read"
			}
			`,
		}
		wm, err := ap.Upsert(policy, nil)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if wm.LastIndex == 0 {
			t.Fatalf("bad index: %d", wm.LastIndex)
		}

		// Get token for read policy
		at := client.ACLTokens()
		token := &nomad.ACLToken{
			Name:     "test",
			Type:     "client",
			Policies: []string{"test"},
		}
		out, wm, err := at.Create(token, nil)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		if wm.LastIndex == 0 {
			t.Fatalf("bad index: %d", wm.LastIndex)
		}
		t.Log("client token:", out.SecretID)
		t.Log("client policies:", out.Policies)

		jsonStruct["token"] = out.SecretID
	}

	jsonBody, _ := json.Marshal(jsonStruct)
	req := httptest.NewRequest("GET", "/", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", ContentType)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(nomadnodestatus.List)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	fmt.Printf("response body: \n%s\n", rr.Body)

}
