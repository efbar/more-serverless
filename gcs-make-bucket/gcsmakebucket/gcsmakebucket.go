package gcsmakebucket

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	storage "cloud.google.com/go/storage"
	message "github.com/efbar/more-serverless/slack-message/slackmessage"
	option "google.golang.org/api/option"
)

type RequestBody struct {
	Name                     string            `json:"name"`
	Labels                   map[string]string `json:"labels,omitempty"`
	UniformBucketLevelAccess bool              `json:"uniformBucketLevelAccess,omitempty"`
	Class                    string            `json:"class,omitempty"`
	VersioningEnabled        bool              `json:"versioningEnabled,omitempty"`
	Location                 string            `json:"location,omitempty"`
	LocationType             string            `json:"locationType,omitempty"`
	JsonKeyPath              string            `json:"jsonKeyPath,omitempty"`
	SlackToken               string            `json:"slackToken,omitempty"`
	SlackChannel             string            `json:"slackChannel,omitempty"`
	SlackEmoji               string            `json:"slackEmoji,omitempty"`
}

type Payload struct {
	Name            string `json:"name"`
	ProjectId       string `json:"projectId"`
	GsUri           string `json:"gsUri"`
	CloudConsoleUri string `json:"cloudConsoleUri"`
}

type Response struct {
	Payload Payload             `json:"payload"`
	Headers map[string][]string `json:"headers"`
}

func Serve(w http.ResponseWriter, r *http.Request) {
	var input []byte

	if r.Body != nil {
		defer r.Body.Close()

		body, err := ioutil.ReadAll(r.Body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		input = body
	}

	if len(input) == 0 {
		fmt.Println("empty body")
		fmt.Println("Json parsing error: empty body")
		http.Error(w, "Input data error", http.StatusBadRequest)
		return
	}

	projectId := os.Getenv("PROJECT_ID")
	if len(projectId) == 0 {
		fmt.Println("empty PROJECT_ID")
		http.Error(w, "Input data error", http.StatusBadRequest)
		return
	}
	googleCreds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	rb := RequestBody{}
	err := json.Unmarshal(input, &rb)
	if err != nil {
		fmt.Println("Json parsing error:", err.Error())
		http.Error(w, "Input data error", http.StatusBadRequest)
		return
	}

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

	var storageClient *storage.Client
	ctx := context.Background()

	if len(jsonKey) != 0 {
		storageClient, err = storage.NewClient(ctx, option.WithCredentialsFile(rb.JsonKeyPath))
	} else {
		fmt.Println("jsonkey empty, not using creds option")
		storageClient, err = storage.NewClient(ctx)
	}
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer storageClient.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	bkt := storageClient.Bucket(rb.Name)

	attrs := &storage.BucketAttrs{
		UniformBucketLevelAccess: storage.UniformBucketLevelAccess{
			Enabled: rb.UniformBucketLevelAccess,
		},
		Location:          rb.Location,
		StorageClass:      rb.Class,
		VersioningEnabled: rb.VersioningEnabled,
		Labels:            rb.Labels,
		LocationType:      rb.LocationType,
	}

	if err := bkt.Create(ctx, projectId, attrs); err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	attrs, err = bkt.Attrs(ctx)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	resBody := fmt.Sprintf("Bucket %s created under %s project, gsUri: gs://%s, CloudConsoleUri: https://storage.cloud.google.com/%s\n", attrs.Name, projectId, attrs.Name, attrs.Name)
	if r.Header.Get("Content-Type") == "text/plain" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(resBody))
	} else {

		w.Header().Set("Content-Type", "application/json")
		jsonResponse := Response{
			Payload: Payload{
				Name:            attrs.Name,
				ProjectId:       projectId,
				GsUri:           fmt.Sprintf("gs://%s", attrs.Name),
				CloudConsoleUri: fmt.Sprintf("https://storage.cloud.google.com/%s", attrs.Name),
			},
			Headers: r.Header,
		}
		res, err := json.Marshal(jsonResponse)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(res)
	}
	if len(rb.SlackToken) > 0 && len(rb.SlackChannel) > 0 {
		slackNotification(&rb, resBody)
	}

}

func slackNotification(rb *RequestBody, resBody string) {
	slackToken := rb.SlackToken
	slackChannelID := rb.SlackChannel
	slackEmoji := rb.SlackEmoji
	slackMessage := "GCP's message " + slackEmoji + "\n```" + resBody + "```"

	sent, err := message.Send(slackToken, slackMessage, slackChannelID)
	if err != nil {
		fmt.Printf("slack error: %s\n", err)
	}
	fmt.Println(sent)
}
