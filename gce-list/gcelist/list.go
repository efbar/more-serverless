package list

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	message "github.com/efbar/more-serverless/slack-message/slackmessage"
	"github.com/ryanuber/columnize"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

type Instance struct {
	Name        string `json:"name"`
	Zone        string `json:"zone"`
	MachineType string `json:"machine_type"`
	Preemptible string `json:"preemptible"`
	InternalIP  string `json:"internal_ip"`
	ExternalIP  string `json:"external_ip"`
	Status      string `json:"status"`
}

type RequestBody struct {
	ProjectId    string `json:"projectId"`
	Region       string `json:"region"`
	JsonKeyPath  string `json:"jsonKeyPath,omitempty"`
	SlackToken   string `json:"slackToken,omitempty"`
	SlackChannel string `json:"slackChannel,omitempty"`
	SlackEmoji   string `json:"slackEmoji,omitempty"`
}

type Response struct {
	Payload []Instance          `json:"payload"`
	Headers map[string][]string `json:"headers"`
}

func Serve(w http.ResponseWriter, r *http.Request) {

	var input []byte

	if r.Body != nil {
		defer r.Body.Close()

		reqBody, err := ioutil.ReadAll(r.Body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		input = reqBody
	}

	if len(input) == 0 {
		fmt.Println("empty body")
	}

	rb := RequestBody{}
	err := json.Unmarshal(input, &rb)
	if err != nil {
		fmt.Println("Json parsing error:", err.Error())
		http.Error(w, "Input data error", http.StatusBadRequest)
		return
	}

	projectId := rb.ProjectId
	if len(projectId) == 0 {
		fmt.Println("empty projectId value")
		http.Error(w, "Input data error", http.StatusBadRequest)
		return
	}
	projectRegion := rb.Region
	if len(projectRegion) == 0 {
		fmt.Println("empty region value")
		http.Error(w, "Input data error", http.StatusBadRequest)
		return
	}
	googleCreds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	var serviceAccountPath string
	if len(rb.JsonKeyPath) != 0 {
		serviceAccountPath = rb.JsonKeyPath
		fmt.Printf("using jsonKeyPath value %s\n", rb.JsonKeyPath)
	} else {
		serviceAccountPath = googleCreds
		fmt.Printf("using google creds environment value %s\n", googleCreds)
	}

	var jsonKey string
	if _, err := os.Stat(serviceAccountPath); err == nil {
		jsonKeyFile, _ := ioutil.ReadFile(serviceAccountPath)
		jsonKey = string(jsonKeyFile)
	}

	var computeService *compute.Service
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	if len(jsonKey) != 0 {
		computeService, err = compute.NewService(ctx, option.WithCredentialsFile(rb.JsonKeyPath))
	} else {
		fmt.Println("jsonkey empty, not using creds option")
		computeService, err = compute.NewService(ctx)
	}
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	region, err := computeService.Regions.Get(projectId, projectRegion).Do()
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Content-Type") == "text/plain" {
		out := []string{"NAME\tZONE\tMACHINE_TYPE\tPREEMPTIBLE\tINTERNAL_IP\tEXTERNAL_IP\tSTATUS"}
		for _, val := range region.Zones {
			instances, _ := computeService.Instances.List(projectId, val[strings.LastIndex(val, "/")+1:]).Do()
			for _, v := range instances.Items {
				zones := strings.Split(v.Zone, "/")
				mType := strings.Split(v.MachineType, "/")
				out = append(out, v.Name+"\t"+zones[len(zones)-1]+"\t"+mType[len(mType)-1]+"\t"+strconv.FormatBool(v.Scheduling.Preemptible)+"\t"+v.NetworkInterfaces[0].NetworkIP+"\t"+v.NetworkInterfaces[0].AccessConfigs[0].NatIP+"\t"+v.Status)
			}
		}
		columnConf := columnize.DefaultConfig()
		columnConf.Delim = "\t"
		columnConf.Glue = "  "
		columnConf.NoTrim = false
		resBody := columnize.Format(out, columnConf)

		if len(rb.SlackToken) > 0 && len(rb.SlackChannel) > 0 {
			slackNotification(&rb, resBody)
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-type", "text/plain")
		w.Write([]byte(resBody))
	} else {
		var vmList []Instance
		for _, val := range region.Zones {
			var vm Instance
			instances, _ := computeService.Instances.List(projectId, val[strings.LastIndex(val, "/")+1:]).Do()
			for _, v := range instances.Items {
				zones := strings.Split(v.Zone, "/")
				mType := strings.Split(v.MachineType, "/")
				vm.Name = v.Name
				vm.Zone = zones[len(zones)-1]
				vm.MachineType = mType[len(mType)-1]
				vm.Preemptible = strconv.FormatBool(v.Scheduling.Preemptible)
				vm.InternalIP = v.NetworkInterfaces[0].NetworkIP
				vm.ExternalIP = v.NetworkInterfaces[0].AccessConfigs[0].NatIP
				vm.Status = v.Status
				vmList = append(vmList, vm)
			}
		}
		jsonResponse := Response{
			Payload: vmList,
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

func slackNotification(rb *RequestBody, resBody string) {
	slackToken := rb.SlackToken
	slackChannelID := rb.SlackChannel
	slackEmoji := rb.SlackEmoji
	slackMessage := "GCP message " + slackEmoji + "\n```" + resBody + "```"

	sent, err := message.Send(slackToken, slackMessage, slackChannelID)
	if err != nil {
		fmt.Printf("slack error: %s\n", err)
	}
	fmt.Println(sent)
}
