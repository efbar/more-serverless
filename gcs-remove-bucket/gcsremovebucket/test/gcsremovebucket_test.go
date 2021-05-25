package test

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gcsremovebucket "github.com/efbar/more-serverless/gcs-remove-bucket/gcsremovebucket"
)

var jsonData string

func init() {
	flag.StringVar(&jsonData, "jsonData", "", "json data input")
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

		req := httptest.NewRequest("GET", "/", strings.NewReader(jsonData))
		req.Header.Set("Content-Type", tr.contentType)

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(gcsremovebucket.Serve)

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		fmt.Printf("response body: \n%s\n", rr.Body)
	}
}
