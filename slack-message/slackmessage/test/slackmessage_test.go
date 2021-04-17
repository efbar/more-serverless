package test

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	message "github.com/efbar/more-serverless/slack-message/slackmessage"
)

var jsonString string

func init() {
	flag.StringVar(&jsonString, "jsonstring", "", "json input string")
}

func TestFunc(t *testing.T) {

	tt := []struct {
		contentType string
	}{
		{
			contentType: "text/plain",
		},
	}

	for _, tr := range tt {

		req := httptest.NewRequest("GET", "/", strings.NewReader(string(jsonString)))
		req.Header.Set("Content-Type", tr.contentType)

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(message.Serve)

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		fmt.Printf("response body: \n%s\n", rr.Body)

	}
}
