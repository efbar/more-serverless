package vaultread

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

// if secret.LeaseDuration > 0 {
// 	if secret.LeaseID != "" {
// 		out = append(out, fmt.Sprintf("lease_id %s\t", secret.LeaseID))
// 		out = append(out, fmt.Sprintf("lease_duration %v\t", time.Duration(secret.LeaseDuration)))
// 		out = append(out, fmt.Sprintf("lease_renewable %t\t", secret.Renewable))
// 	} else {
// 		out = append(out, fmt.Sprintf("refresh_interval %v\t", time.Duration(secret.LeaseDuration)))
// 	}
// }

// if secret.Auth != nil {
// 	out = append(out, fmt.Sprintf("token %s\t", secret.Auth.ClientToken))
// 	out = append(out, fmt.Sprintf("token_accessor %s\t", secret.Auth.Accessor))
// 	if secret.Auth.LeaseDuration == 0 {
// 		out = append(out, fmt.Sprintf("token_duration %s\t", "∞"))
// 	} else {
// 		out = append(out, fmt.Sprintf("token_duration %v\t", time.Duration(secret.Auth.LeaseDuration)))
// 	}
// 	out = append(out, fmt.Sprintf("token_renewable %t\t", secret.Auth.Renewable))
// 	out = append(out, fmt.Sprintf("token_policies %q\t", secret.Auth.TokenPolicies))
// 	out = append(out, fmt.Sprintf("identity_policies %q\t", secret.Auth.IdentityPolicies))
// 	out = append(out, fmt.Sprintf("policies %q\t", secret.Auth.Policies))
// 	for k, v := range secret.Auth.Metadata {
// 		out = append(out, fmt.Sprintf("token_meta_%s\t%v", k, v))
// 	}
// }

// if secret.WrapInfo != nil {
// 	out = append(out, fmt.Sprintf("wrapping_token: %s\t", secret.WrapInfo.Token))
// 	out = append(out, fmt.Sprintf("wrapping_accessor: %s\t", secret.WrapInfo.Accessor))
// 	out = append(out, fmt.Sprintf("wrapping_token_ttl: %v\t", time.Duration(secret.WrapInfo.TTL)))
// 	out = append(out, fmt.Sprintf("wrapping_token_creation_time: %s\t", secret.WrapInfo.CreationTime.String()))
// 	out = append(out, fmt.Sprintf("wrapping_token_creation_path: %s\t", secret.WrapInfo.CreationPath))
// 	if secret.WrapInfo.WrappedAccessor != "" {
// 		out = append(out, fmt.Sprintf("wrapped_accessor: %s\t", secret.WrapInfo.WrappedAccessor))
// 	}
// }
