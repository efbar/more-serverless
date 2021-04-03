package function

import (
	"net/http"

	vaultstatus "github.com/efbar/more-serverless/vault-status/vaultstatus"
)

func Handle(w http.ResponseWriter, r *http.Request) {

	vaultstatus.Serve(w, r)
}
