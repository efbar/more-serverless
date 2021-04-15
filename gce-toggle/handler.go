package function

import (
	"net/http"

	toggle "github.com/efbar/more-serverless/gce-toggle/gcetoggle"
)

func Handle(w http.ResponseWriter, r *http.Request) {

	toggle.Serve(w, r)

}
