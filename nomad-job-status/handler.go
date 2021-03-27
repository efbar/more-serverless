package function

import (
	"net/http"

	nomadjobstatus "github.com/efbar/more-serverless/nomad-job-status/nomadjobstatus"
)

func Handle(w http.ResponseWriter, r *http.Request) {

	nomadjobstatus.List(w, r)

}
