package function

import (
	"net/http"

	vaulttransit "github.com/efbar/more-serverless/vault-transit/vaulttransit"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	vaulttransit.Serve(w, r)
}
