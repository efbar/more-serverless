package function

import (
	"net/http"

	consulcatalogservices "github.com/efbar/more-serverless/consul-catalog-services/consulcatalogservices"
)

func Handle(w http.ResponseWriter, r *http.Request) {

	consulcatalogservices.Serve(w, r)
}
