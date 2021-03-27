package function

import (
	"net/http"

	nomadnodes "github.com/efbar/more-serverless/nomad-nodes-list/nomadnodes"
)

func Handle(w http.ResponseWriter, r *http.Request) {

	nomadnodes.List(w, r)

}
