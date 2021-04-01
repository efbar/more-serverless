package function

import (
	"net/http"

	nomadservermembers "github.com/efbar/more-serverless/nomad-server-members/nomadservermembers"
)

func Handle(w http.ResponseWriter, r *http.Request) {

	nomadservermembers.Serve(w, r)
}
