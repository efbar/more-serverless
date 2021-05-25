package function

import (
	"net/http"

	"github.com/efbar/more-serverless/gcs-remove-bucket/gcsremovebucket"
)

func Handle(w http.ResponseWriter, r *http.Request) {
	gcsremovebucket.Serve(w, r)
}
