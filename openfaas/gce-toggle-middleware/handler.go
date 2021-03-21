package function

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"../../lib/gcp/gce/toggle"
)

func Handle(w http.ResponseWriter, r *http.Request) {
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

	fmt.Printf("request body: %s", string(input))

	toggle.Toggle(w, r)

}
