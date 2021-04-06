package function

import (
	"net/http"

	"github.com/efbar/more-serverless/vault-kv-put/vaultkvput"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	vaultkvput.Serve(w, r)
}
