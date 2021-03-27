package nomadjobstatus

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/ryanuber/columnize"

	nomad "github.com/hashicorp/nomad/api"
)

type RequestBody struct {
	Token    string `json:"token"`
	Endpoint string `json:"endpoint"`
}

type Response struct {
	Payload     []Job               `json:"payload"`
	Headers     map[string][]string `json:"headers"`
	Environment []string            `json:"environment"`
}

type Job struct {
	ID         *string `json:"id"`
	Name       *string `json:"name"`
	Type       *string `json:"type"`
	Priority   *int    `json:"priority"`
	Status     *string `json:"status"`
	SubmitTime *int64  `json:"submitTime"`
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

	jobs := client.Jobs()

	resp, _, err := jobs.List(nil)
	if err != nil {
		fmt.Printf("listing jobs error: %#v", resp)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resBody := ""
	if r.Header.Get("Content-Type") == "text/plain" {
		out := []string{"ID\tType\tPriority\tStatus\tSubmitTime"}
		for _, v := range resp {
			out = append(out, v.ID+"\t"+v.Type+"\t"+fmt.Sprint(v.Priority)+"\t"+v.Status+"\t"+time.Unix(0, v.SubmitTime).Format("2006-01-02T15:04:05Z07:00"))
		}
		columnConf := columnize.DefaultConfig()
		columnConf.Delim = "\t"
		columnConf.Glue = "  "
		columnConf.NoTrim = false
		resBody = columnize.Format(out, columnConf)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(resBody))
	} else {
		var jobList []Job
		for _, v := range resp {
			job := Job{
				Status:     &v.Status,
				ID:         &v.ID,
				Name:       &v.Name,
				Priority:   &v.Priority,
				SubmitTime: &v.SubmitTime,
				Type:       &v.Type,
			}
			jobList = append(jobList, job)
		}
		jsonResponse := Response{
			Payload:     jobList,
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
