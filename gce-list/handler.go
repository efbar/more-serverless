package function

import (
	"net/http"

	list "github.com/efbar/more-serverless/gce-list/gcelist"
)

func Handle(w http.ResponseWriter, r *http.Request) {

	list.Serve(w, r)

}
