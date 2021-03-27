package nomadnodestatus

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/ryanuber/columnize"

	nomad "github.com/hashicorp/nomad/api"
)

type RequestBody struct {
	Token    string `json:"token"`
	Endpoint string `json:"endpoint"`
}

type Response struct {
	Payload     []Node              `json:"payload"`
	Headers     map[string][]string `json:"headers"`
	Environment []string            `json:"environment"`
}

type Node struct {
	ID                    *string `json:"id"`
	Datacenter            *string `json:"datacenter"`
	Name                  *string `json:"name"`
	NodeClass             *string `json:"nodeclass"`
	Drain                 *bool   `json:"drain,omitempty"`
	SchedulingEligibility *string `json:"scheduling-eligibility"`
	Status                *string `json:"status"`
}

func List(w http.ResponseWriter, r *http.Request) {
	var input []byte

	if r.Body != nil {
		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)

		input = body
	}

	if len(input) != 0 {
		fmt.Printf("request body: %s\n", string(input))
	} else {
		fmt.Println("empty body")
	}

	rb := RequestBody{}
	json.Unmarshal(input, &rb)

	nomadEndpoint := "http://localhost:4646"
	if rb.Endpoint != "" {
		nomadEndpoint = rb.Endpoint
	}

	conf := nomad.DefaultConfig()

	conf.Address = nomadEndpoint

	client, err := nomad.NewClient(conf)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rb.Token != "" {
		client.SetSecretID(rb.Token)
	}

	nodes := client.Nodes()

	resp, _, err := nodes.List(nil)
	if err != nil {
		fmt.Printf("listing nodes error: %#v", resp)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resBody := ""
	if r.Header.Get("Content-Type") == "text/plain" {
		out := []string{"ID\tDC\tName\tClass\tDrain\tEligibility\tStatus"}
		for _, v := range resp {
			if len(v.NodeClass) == 0 {
				v.NodeClass = "<none>"
			}
			out = append(out, v.ID+"\t"+v.Datacenter+"\t"+v.Name+"\t"+v.NodeClass+"\t"+strconv.FormatBool(v.Drain)+"\t"+v.SchedulingEligibility+"\t"+v.Status)
		}
		columnConf := columnize.DefaultConfig()
		columnConf.Delim = "\t"
		columnConf.Glue = "  "
		columnConf.NoTrim = false
		resBody = columnize.Format(out, columnConf)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(resBody))
	} else {
		var nodeList []Node
		for _, v := range resp {
			node := Node{
				ID:                    &v.ID,
				Datacenter:            &v.Datacenter,
				Name:                  &v.Name,
				NodeClass:             &v.NodeClass,
				Drain:                 &v.Drain,
				SchedulingEligibility: &v.SchedulingEligibility,
				Status:                &v.Status,
			}
			nodeList = append(nodeList, node)
		}
		jsonResponse := Response{
			Payload:     nodeList,
			Headers:     r.Header,
			Environment: os.Environ(),
		}
		resBody, err := json.Marshal(jsonResponse)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(resBody)

	}

	w.WriteHeader(http.StatusOK)

}
