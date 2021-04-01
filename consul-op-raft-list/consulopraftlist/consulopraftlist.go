package consulopraftlist

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"

	consul "github.com/hashicorp/consul/api"
	"github.com/ryanuber/columnize"
)

type RequestBody struct {
	Token    string `json:"token"`
	Endpoint string `json:"endpoint"`
}

type Response struct {
	Payload []Peer              `json:"payload"`
	Headers map[string][]string `json:"headers"`
}

type Peer struct {
	Node         *string `json:"node"`
	ID           *string `json:"id"`
	Address      *string `json:"address"`
	State        string  `json:"state"`
	Voter        *bool   `json:"voter"`
	RaftProtocol string  `json:"raft-protocol"`
}

func Serve(w http.ResponseWriter, r *http.Request) {

	var input []byte

	if r.Body != nil {
		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)

		input = body
	}

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

	q := &consul.QueryOptions{
		AllowStale: true,
	}
	peers, err := client.Operator().RaftGetConfiguration(q)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Content-Type") == "text/plain" {
		out := []string{"Node\tID\tAddress\tState\tVoter\tRaftProtocol"}
		for _, v := range peers.Servers {
			raftProtocol, state := getInfo(v)
			out = append(
				out,
				fmt.Sprintf("%s\t%s\t%s\t%s\t%v\t%s",
					v.Node, v.ID, v.Address, state, v.Voter, raftProtocol))
		}
		resBody := ""
		sort.Strings(out)
		columnConf := columnize.DefaultConfig()
		columnConf.Delim = "\t"
		columnConf.Glue = "  "
		columnConf.NoTrim = false
		resBody = columnize.Format(out, columnConf)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(resBody))
	} else {
		var peersList []Peer
		for _, v := range peers.Servers {
			raftProtocol, state := getInfo(v)
			peer := Peer{
				Node:         &v.Node,
				ID:           &v.ID,
				Address:      &v.Address,
				State:        state,
				Voter:        &v.Voter,
				RaftProtocol: raftProtocol,
			}
			peersList = append(peersList, peer)
		}
		jsonResponse := Response{
			Payload: peersList,
			Headers: r.Header,
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

func getInfo(v *consul.RaftServer) (string, string) {
	raftProtocol := v.ProtocolVersion
	if raftProtocol == "" {
		raftProtocol = "<=1"
	}
	state := "follower"
	if v.Leader {
		state = "leader"
	}

	return raftProtocol, state
}
