package consulcatalogservices

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"

	consul "github.com/hashicorp/consul/api"
	"github.com/ryanuber/columnize"
)

type RequestBody struct {
	Token    string `json:"token"`
	Endpoint string `json:"endpoint"`
}

type Response struct {
	Payload []SimpleService     `json:"payload"`
	Headers map[string][]string `json:"headers"`
}

type SimpleService struct {
	Name string   `json:"name"`
	Tags []string `json:"tags,omitempty"`
}

func Serve(w http.ResponseWriter, r *http.Request) {
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

	catalog := client.Catalog()

	services, _, _ := catalog.Services(nil)

	resBody := ""
	if r.Header.Get("Content-Type") == "text/plain" {
		var out []string
		for k, v := range services {
			fmt.Println(k, v)
			out = append(out, k+"\t"+strings.Join(v, ","))
		}
		sort.Strings(out)
		columnConf := columnize.DefaultConfig()
		columnConf.Delim = "\t"
		columnConf.Glue = "      "
		columnConf.NoTrim = false
		resBody = columnize.Format(out, columnConf)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(resBody))
	} else {
		var simpleServiceList []SimpleService
		for k, v := range services {
			simpleService := SimpleService{
				Name: k,
				Tags: v,
			}
			simpleServiceList = append(simpleServiceList, simpleService)
		}
		jsonResponse := Response{
			Payload: simpleServiceList,
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
