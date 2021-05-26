package gcscpbucket

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	storage "cloud.google.com/go/storage"
	message "github.com/efbar/more-serverless/slack-message/slackmessage"
	iterator "google.golang.org/api/iterator"
	option "google.golang.org/api/option"
)

type RequestBody struct {
	SrcBucket    string `json:"srcBucket"`
	DstBucket    string `json:"dstBucket"`
	ProjectId    string `json:"projectId"`
	JsonKeyPath  string `json:"jsonKeyPath,omitempty"`
	SlackToken   string `json:"slackToken,omitempty"`
	SlackChannel string `json:"slackChannel,omitempty"`
	SlackEmoji   string `json:"slackEmoji,omitempty"`
}

type Result struct {
	SrcObj    string `json:srcObj`
	Completed bool   `json:completed`
	Error     string `json:"error,omitempty"`
}

type Response struct {
	Result  []Result            `json:"result"`
	Total   int64               `json:"totalSize"`
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
	srcBucket := rb.SrcBucket
	if len(srcBucket) == 0 {
		fmt.Println("empty projectId value")
		http.Error(w, "Input data error", http.StatusBadRequest)
		return
	}
	dstBucket := rb.DstBucket
	if len(dstBucket) == 0 {
		fmt.Println("empty projectId value")
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

	srcBkt := storageClient.Bucket(rb.SrcBucket).Objects(ctx, nil)

	var srcObjList []Result
	var totSize int64
	totNumber := 0

	for {
		totNumber = totNumber + 1
		attrs, err := srcBkt.Next()

		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Printf("Bucket(%q).Objects: %v", srcBucket, err)
		}

		totSize = totSize + attrs.Size

		dstObjName := attrs.Name
		srcObj := storageClient.Bucket(rb.SrcBucket).Object(attrs.Name)
		dstObj := storageClient.Bucket(rb.DstBucket).Object(dstObjName)
		res := Result{
			SrcObj:    attrs.Name,
			Completed: true,
			Error:     "",
		}
		if _, err := dstObj.CopierFrom(srcObj).Run(ctx); err != nil {
			res = Result{
				SrcObj:    attrs.Name,
				Completed: false,
				Error:     err.Error(),
			}
			fmt.Printf("Object(%v).CopierFrom(%v).Run error: %v", dstObj, srcObj, err)
		}
		srcObjList = append(srcObjList, res)
	}

	const iecKib = 1024
	var totSizeString string
	if totSize >= iecKib {
		div, exp := int64(iecKib), 0 // this cool part taken from here: https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
		for n := totSize / iecKib; n >= iecKib; n /= iecKib {
			div *= iecKib
			exp++
		}
		totSizeString = fmt.Sprintf("%.1f %ciB", float64(totSize)/float64(div), "KMGTPE"[exp])
	} else {
		totSizeString = strconv.Itoa(int(totSize))
	}

	w.WriteHeader(http.StatusOK)
	resBody := fmt.Sprintf("Operation completed over %d objects/%s.\n", totNumber, totSizeString)
	if r.Header.Get("Content-Type") == "text/plain" {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(resBody))
	} else {

		w.Header().Set("Content-Type", "application/json")
		jsonResponse := Response{
			Result:  srcObjList,
			Total:   totSize,
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
