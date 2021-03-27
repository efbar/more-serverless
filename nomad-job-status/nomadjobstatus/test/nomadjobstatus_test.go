package testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	nomadjobstatus "github.com/efbar/more-serverless/nomad-job-status/nomadjobstatus"
	nomad "github.com/hashicorp/nomad/api"
	testutil "github.com/hashicorp/nomad/testutil"
)

func TestFunc(t *testing.T) {

	// Create server
	conf := nomad.DefaultConfig()

	ACLEnabled := false
	Loglevel := "INFO"
	ContentType := "text/plain"

	server := testutil.NewTestServer(t, func(c *testutil.TestServerConfig) {
		c.ACL.Enabled = ACLEnabled
		c.LogLevel = Loglevel
	})
	defer server.Stop()
	conf.Address = "http://" + server.HTTPAddr

	// Create client
	client, err := nomad.NewClient(conf)
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

	jobs := client.Jobs()
	job := testJob()
	resp, wm, err := jobs.Register(job, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if wm.LastIndex == 0 {
		t.Fatalf("bad index: %d", wm.LastIndex)
	}
	if resp == nil {
		t.Fatalf("job not registered")
	}
	if len(resp.EvalID) == 0 {
		t.Fatalf("job not evaluated")
	}

	jsonBody, _ := json.Marshal(jsonStruct)
	req := httptest.NewRequest("GET", "/", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", ContentType)

	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(nomadjobstatus.List)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	fmt.Printf("response body: \n%s\n", rr.Body)

}

func testJob() *nomad.Job {
	task := nomad.NewTask("task1", "raw_exec").
		SetConfig("command", "/bin/sleep").
		Require(&nomad.Resources{
			CPU:      intToPtr(100),
			MemoryMB: intToPtr(256),
		}).
		SetLogConfig(&nomad.LogConfig{
			MaxFiles:      intToPtr(1),
			MaxFileSizeMB: intToPtr(2),
		})

	group := nomad.NewTaskGroup("group1", 1).
		AddTask(task).
		RequireDisk(&nomad.EphemeralDisk{
			SizeMB: intToPtr(25),
		})

	job := nomad.NewBatchJob("job1", "redis", "global", 1).
		AddDatacenter("dc1").
		AddTaskGroup(group)

	return job
}

func intToPtr(i int) *int {
	return &i
}
