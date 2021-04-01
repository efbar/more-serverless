package function

import (
	"net/http"

	nomadnodestatus "github.com/efbar/more-serverless/nomad-node-status/nomadnodestatus"
)

func Handle(w http.ResponseWriter, r *http.Request) {

	nomadnodestatus.Serve(w, r)

}
