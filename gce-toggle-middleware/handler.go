package function

import (
	"net/http"

	"github.com/efbar/more-serverless/google/toggle"
)

func Handle(w http.ResponseWriter, r *http.Request) {

	toggle.Toggle(w, r)

}
