package testing

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	toggle "github.com/efbar/more-serverless/gce-toggle/gcetoggle"
)

var jsonkeyPath string

func init() {
	flag.StringVar(&jsonkeyPath, "jsonkey", "", "json key file")
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

		jkey, err := ioutil.ReadFile(jsonkeyPath)
		if err != nil {
			t.Log(err.Error())
		}

		req := httptest.NewRequest("GET", "/", strings.NewReader(string(jkey)))
		req.Header.Set("Content-Type", tr.contentType)

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(toggle.Serve)

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		fmt.Printf("response body: \n%s\n", rr.Body)
	}
}
