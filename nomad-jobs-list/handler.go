package function

import (
	"net/http"

	nomadjobs "github.com/efbar/more-serverless/nomad-list-jobs/nomadjobs"
)

func Handle(w http.ResponseWriter, r *http.Request) {

	nomadjobs.List(w, r)

}
