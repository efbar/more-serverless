package function

import (
	"net/http"

	consulmembers "github.com/efbar/more-serverless/consul-members/consulmembers"
)

func Handle(w http.ResponseWriter, r *http.Request) {

	consulmembers.List(w, r)
}
