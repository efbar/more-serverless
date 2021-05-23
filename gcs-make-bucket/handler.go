package function

import (
	"net/http"

	"github.com/efbar/more-serverless/gcs-make-bucket/gcsmakebucket"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	gcsmakebucket.Serve(w, r)
}
