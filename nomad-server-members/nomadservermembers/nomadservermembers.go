package nomadservermembers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"

	"github.com/ryanuber/columnize"

	nomad "github.com/hashicorp/nomad/api"
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
	Name        *string `json:"name"`
	Addr        *string `json:"addr"`
	Port        *uint16 `json:"port"`
	Status      *string `json:"status"`
	Leader      string  `json:"leader"`
	ProtocolCur *uint8  `json:"protocol"`
	Build       string  `json:"build"`
	Datacenter  string  `json:"datacenter"`
	Region      string  `json:"region"`
}

func Serve(w http.ResponseWriter, r *http.Request) {
	var input []byte

	if r.Body != nil {
		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)

		input = body
	}

	rb := RequestBody{}
	err := json.Unmarshal(input, &rb)
	if err != nil {
		fmt.Println("Json parsing error:", err.Error())
		http.Error(w, "Input data error", http.StatusBadRequest)
		return
	}

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

	agent := client.Agent()
	leader, _ := client.Status().Leader()

	m, err := agent.Members()
	if err != nil {
		fmt.Printf("err: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sort.Sort(nomad.AgentMembersNameSort(m.Members))

	resBody := ""
	if r.Header.Get("Content-Type") == "text/plain" {
		out := []string{"Name\tAddress\tPort\tStatus\tLeader\tProtocol\tBuild\tDatacenter\tRegion"}
		for _, v := range m.Members {
			out = append(out, v.Name+"\t"+v.Addr+"\t"+strconv.Itoa(int(v.Port))+"\t"+v.Status+"\t"+isLeader(leader, v)+"\t"+strconv.Itoa(int(v.ProtocolCur))+"\t"+v.Tags["build"]+"\t"+v.Tags["dc"]+"\t"+v.Tags["region"])
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
		for _, v := range m.Members {
			member := AgentMember{
				Name:        &v.Name,
				Addr:        &v.Addr,
				Port:        &v.Port,
				Status:      &v.Status,
				Leader:      isLeader(leader, v),
				ProtocolCur: &v.ProtocolCur,
				Build:       v.Tags["build"],
				Datacenter:  v.Tags["dc"],
				Region:      v.Tags["region"],
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

	w.WriteHeader(http.StatusOK)

}

func isLeader(leader string, v *nomad.AgentMember) string {
	isLeader := false
	if leader == net.JoinHostPort(v.Addr, v.Tags["port"]) {
		isLeader = true
	}
	return strconv.FormatBool(isLeader)
}
