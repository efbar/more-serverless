package function

import (
	"net/http"

	consulopraftlist "github.com/efbar/more-serverless/consul-op-raft-list/consulopraftlist"
)

func Handle(w http.ResponseWriter, r *http.Request) {

	consulopraftlist.List(w, r)

}
