package function

import (
	"net/http"

	"github.com/efbar/more-serverless/vault-kv-get/vaultkvget"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	vaultkvget.Serve(w, r)
}
