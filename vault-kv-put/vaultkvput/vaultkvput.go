package vaultkvput

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	vault "github.com/hashicorp/vault/api"
	"github.com/ryanuber/columnize"
)

type RequestBody struct {
	Token    string                 `json:"token"`
	Endpoint string                 `json:"endpoint"`
	Path     string                 `json:"path"`
	Data     map[string]interface{} `json:"data"`
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

	queryData := map[string]interface{}{
		"data":    data,
		"options": map[string]interface{}{},
	}

	var secret *vault.Secret
	if len(data) == 0 {
		fmt.Println("no data")
		http.Error(w, "no data", http.StatusInternalServerError)
		return
	}
	secret, err = client.Logical().Write(path, queryData)
	if err != nil {
		fmt.Printf("Error writing data to %s: %s", path, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resData := secret.Data

	resBody := ""
	if r.Header.Get("Content-Type") == "text/plain" {
		out := []string{}
		out = append(out, "Key\tValue")
		out = append(out, "---\t-----")
		out = append(out, formatData(resData)...)
		out = append(out, "")

		columnConf := columnize.DefaultConfig()
		columnConf.Delim = "\t"
		columnConf.Glue = ""
		columnConf.NoTrim = false
		resBody = columnize.Format(out, columnConf)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(resBody))
	} else {
		jsonResponse := Response{
			Payload: resData,
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
