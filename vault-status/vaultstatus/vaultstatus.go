package vaultstatus

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	vault "github.com/hashicorp/vault/api"
)

type RequestBody struct {
	Token    string `json:"token"`
	Endpoint string `json:"endpoint"`
}

// type Response struct {
// 	Payload []Peer              `json:"payload"`
// 	Headers map[string][]string `json:"headers"`
// }

func Status(w http.ResponseWriter, r *http.Request) {

	var input []byte

	if r.Body != nil {
		defer r.Body.Close()

		body, _ := ioutil.ReadAll(r.Body)

		input = body
	}

	rb := RequestBody{}
	json.Unmarshal(input, &rb)

	vaultEndpoint := "http://localhost:8500"
	if rb.Endpoint != "" {
		vaultEndpoint = rb.Endpoint
	}

	conf := vault.DefaultConfig()

	conf.Address = vaultEndpoint

	if rb.Token != "" {
		conf.Token = rb.Token
	}

}
