package function

import (
	"net/http"

	"github.com/efbar/more-serverless/gcs-cp-bucket/gcscpbucket"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	gcscpbucket.Serve(w, r)
}
