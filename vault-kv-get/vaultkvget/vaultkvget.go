package vaultkvget

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	vault "github.com/hashicorp/vault/api"
	"github.com/ryanuber/columnize"
)

type RequestBody struct {
	Token    string              `json:"token"`
	Endpoint string              `json:"endpoint"`
	Path     string              `json:"path"`
	Data     map[string][]string `json:"data"`
}

type Response struct {
	Payload map[string]interface{} `json:"payload"`
	Headers map[string][]string    `json:"headers"`
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

	vaultEndpoint := "http://localhost:8200"
	if rb.Endpoint != "" {
		vaultEndpoint = rb.Endpoint
	}

	conf := vault.DefaultConfig()

	conf.Address = vaultEndpoint

	client, err := vault.NewClient(conf)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if rb.Token != "" {
		client.SetToken(rb.Token)
	}

	path := rb.Path
	if len(path) == 0 {
		fmt.Println("empty path")
		http.Error(w, "empty path", http.StatusInternalServerError)
		return
	}

	data := rb.Data
	var secret *vault.Secret
	if len(data) == 0 {
		secret, err = client.Logical().ReadWithData(path, nil)
	} else {
		secret, err = client.Logical().ReadWithData(path, data)
	}
	if err != nil {
		fmt.Println("Reading error")
		http.Error(w, "Reading error", http.StatusInternalServerError)
		return
	}

	resBody := ""
	if r.Header.Get("Content-Type") == "text/plain" {
		out := []string{}
		realMetadata := secret.Data["metadata"].(map[string]interface{})
		realData := secret.Data["data"].(map[string]interface{})
		out = append(out, "Metadata values:\t")
		out = append(out, "======== ====== \t")
		out = append(out, "Key\tValue")
		out = append(out, "---\t-----")
		out = append(out, formatData(realMetadata)...)
		out = append(out, "")
		out = append(out, "Data values:\t")
		out = append(out, "==== ====== \t")
		out = append(out, "Key\tValue")
		out = append(out, "---\t-----")
		out = append(out, formatData(realData)...)

		columnConf := columnize.DefaultConfig()
		columnConf.Delim = "\t"
		columnConf.Glue = ""
		columnConf.NoTrim = false
		resBody = columnize.Format(out, columnConf)
		fmt.Println(len(resBody))
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(resBody))
	} else {
		jsonResponse := Response{
			Payload: secret.Data,
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

func formatData(rawData map[string]interface{}) []string {
	out := []string{}
	if len(rawData) > 0 {
		for k, v := range rawData {
			out = append(out, fmt.Sprintf("%s\t%v", k, v))
		}
	}
	return out
}
