package function

import (
	"net/http"

	"github.com/efbar/more-serverless/vault-read/vaultread"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	vaultread.Serve(w, r)
}
