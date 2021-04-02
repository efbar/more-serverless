package testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/efbar/more-serverless/consul-members/consulmembers"
	consul "github.com/hashicorp/consul/api"
	testutil "github.com/hashicorp/consul/sdk/testutil"
	"github.com/hashicorp/consul/sdk/testutil/retry"
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
			aclenabled:  false,
			loglevel:    "INFO",
			contentType: "application/json",
		},
	}

	for _, tr := range tt {

		var server *testutil.TestServer
		var err error
		retry.RunWith(retry.ThreeTimes(), t, func(r *retry.R) {
			server, err = testutil.NewTestServerConfigT(t, func(c *testutil.TestServerConfig) {
				c.LogLevel = tr.loglevel
				if tr.aclenabled {
					c.ACL.Enabled = tr.aclenabled
					c.PrimaryDatacenter = "dc1"
					c.ACL.Tokens.Master = "root"
					c.ACL.Tokens.Agent = "root"
					c.ACL.Enabled = true
					c.ACL.DefaultPolicy = "deny"
				}
			})
		})
		if err != nil {
			t.Fatalf("Failed to start server: %v", err.Error())
		}
		defer server.Stop()
		server.WaitForSerfCheck(t)

		if server.Config.Bootstrap {
			server.WaitForLeader(t)
		}

		conf := consul.DefaultConfig()
		conf.Address = "http://" + server.HTTPAddr

		jsonStruct := map[string]interface{}{
			"endpoint": conf.Address,
		}

		// here acl token part
		if server.Config.ACL.Enabled {
			conf.Token = "root"
			client, err := consul.NewClient(conf)
			if err != nil {
				t.Fatalf("acl err: %v", err)
			}

			acl := client.ACL()

			created, _, err := acl.PolicyCreate(&consul.ACLPolicy{
				Name:        "test-policy",
				Description: "test-policy description",
				Rules: `node_prefix "" { 
				policy = "read" 
			}
			agent_prefix "" {
				policy = "read"
			}
			operator = "read"`,
				Datacenters: []string{"dc1"},
			}, nil)
			if err != nil {
				t.Fatalf("policy err: %v", err)
			}

			tokenTest, _, err := acl.TokenCreate(&consul.ACLToken{
				Description: created.Description + " token",
				Policies: []*consul.ACLTokenPolicyLink{
					{
						ID: created.ID,
					},
				},
			}, nil)
			if err != nil {
				t.Fatalf("token creation err: %v", err)
			}

			t.Log("created token:", tokenTest.SecretID)
			t.Log("created policies:", tokenTest.Policies[0].ID)

			jsonStruct["token"] = tokenTest.SecretID

		}

		jsonBody, _ := json.Marshal(jsonStruct)
		req := httptest.NewRequest("GET", "/", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", tr.contentType)

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(consulmembers.Serve)

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		fmt.Printf("response body: \n%s\n", rr.Body)
	}

}
