package message

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/slack-go/slack"
)

type RequestBody struct {
	Token   string `json:"token"`
	Message string `json:"message"`
	Channel string `json:"channel"`
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
		fmt.Println("json parsing error:", err.Error())
		http.Error(w, "Input data error", http.StatusBadRequest)
		return
	}

	if len(rb.Token) == 0 {
		fmt.Println("error: no token")
		http.Error(w, "error, no channel", http.StatusBadRequest)
		return
	}
	if len(rb.Message) == 0 {
		fmt.Println("error: no message")
		http.Error(w, "error, no message", http.StatusBadRequest)
		return
	}
	if len(rb.Channel) == 0 {
		fmt.Println("error: no channel")
		http.Error(w, "error, no channel", http.StatusBadRequest)
		return
	}

	res, err := Send(rb.Token, rb.Message, rb.Channel)
	if err != nil {
		fmt.Printf("slack error: %s\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "text/plain")
	w.Write([]byte(res))

}

func Send(token string, message string, channelID string) (string, error) {
	api := slack.New(token)

	channelID, timestamp, err := api.PostMessage(
		channelID,
		slack.MsgOptionText(message, false),
		slack.MsgOptionAsUser(true),
	)
	if err != nil {
		return "slack message not sent", err
	}
	timest, _ := strconv.ParseFloat(timestamp, 64)
	return fmt.Sprintf("Message successfully sent to channel %s at %s", channelID, time.Unix(int64(timest), 0).Format("2006-01-02T15:04:05Z07:00")), nil
}
