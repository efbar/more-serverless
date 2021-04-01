package consulmembers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/ryanuber/columnize"

	consul "github.com/hashicorp/consul/api"
)

type RequestBody struct {
	Token    string `json:"token"`
	Endpoint string `json:"endpoint"`
}

type Response struct {
	Payload     []AgentMember       `json:"payload"`
	Headers     map[string][]string `json:"headers"`
	Environment []string            `json:"environment"`
}

type AgentMember struct {
	Name        string `json:"name"`
	Addr        string `json:"address"`
	Port        uint16 `json:"port"`
	Status      string `json:"status"`
	Type        string `json:"type"`
	ProtocolCur string `json:"protocol"`
	Build       string `json:"build"`
	Datacenter  string `json:"dc"`
	Segment     string `json:"segment"`
}

type ByMemberNameAndSegment []*consul.AgentMember

func (m ByMemberNameAndSegment) Len() int      { return len(m) }
func (m ByMemberNameAndSegment) Swap(i, j int) { m[i], m[j] = m[j], m[i] }
func (m ByMemberNameAndSegment) Less(i, j int) bool {
	switch {
	case m[i].Tags["segment"] < m[j].Tags["segment"]:
		return true
	case m[i].Tags["segment"] > m[j].Tags["segment"]:
		return false
	default:
		return m[i].Name < m[j].Name
	}
}

func List(w http.ResponseWriter, r *http.Request) {
	var input []byte

	if r.Body != nil {
		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)

		input = body
	}

	// if len(input) != 0 {
	// 	fmt.Printf("request body: %s\n", string(input))
	// } else {
	// 	fmt.Println("empty body")
	// }

	rb := RequestBody{}
	json.Unmarshal(input, &rb)

	consulEndpoint := "http://localhost:8500"
	if rb.Endpoint != "" {
		consulEndpoint = rb.Endpoint
	}

	conf := consul.DefaultConfig()

	conf.Address = consulEndpoint

	if rb.Token != "" {
		conf.Token = rb.Token
	}

	client, err := consul.NewClient(conf)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	agent := client.Agent()
	// Make the request.
	members, err := agent.Members(false)
	if err != nil {
		fmt.Println("error getting members:", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sort.Sort(ByMemberNameAndSegment(members))

	resBody := ""
	if r.Header.Get("Content-Type") == "text/plain" {
		out := []string{"Node\tAddress\tStatus\tType\tBuild\tProtocol\tDC\tSegment"}
		for _, v := range members {
			agentType, agentStatus, build, segments := getMemberInfo(v)
			out = append(out, v.Name+"\t"+v.Addr+"\t"+agentStatus+"\t"+agentType+"\t"+build+"\t"+v.Tags["vsn"]+"\t"+v.Tags["dc"]+"\t"+segments)
		}
		columnConf := columnize.DefaultConfig()
		columnConf.Delim = "\t"
		columnConf.Glue = "  "
		columnConf.NoTrim = false
		resBody = columnize.Format(out, columnConf)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(resBody))
	} else {
		var memberList []AgentMember
		for _, v := range members {
			agentType, agentStatus, build, segments := getMemberInfo(v)
			member := AgentMember{
				Name:        v.Name,
				Addr:        v.Addr,
				Port:        v.Port,
				Status:      agentStatus,
				Type:        agentType,
				ProtocolCur: v.Tags["vsn"],
				Build:       build,
				Datacenter:  v.Tags["dc"],
				Segment:     segments,
			}
			memberList = append(memberList, member)
		}
		jsonResponse := Response{
			Payload:     memberList,
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

}

func getMemberInfo(v *consul.AgentMember) (string, string, string, string) {
	var agentType string
	role := v.Tags["role"]
	switch role {
	case "node":
		agentType = "client"
	case "consul":
		agentType = "server"
	default:
		agentType = "unknown"
	}

	var agentStatus string
	status := v.Status
	switch status {
	case 0:
		agentStatus = "none"
	case 1:
		agentStatus = "alive"
	case 2:
		agentStatus = "leaving"
	case 3:
		agentStatus = "left"
	case 4:
		agentStatus = "failed"
	default:
		agentStatus = "unknown"
	}

	var segment string
	if v.Tags["segment"] == "" && v.Tags["role"] == "consul" {
		segment = "<all>"
	} else {
		segment = "<default>"
	}

	build := v.Tags["build"]
	if build == "" {
		build = "< 0.3"
	} else if idx := strings.Index(build, ":"); idx != -1 {
		build = build[:idx]
	}

	return agentType, agentStatus, build, segment
}
