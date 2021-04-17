package function

import (
	"net/http"

	slackmessage "github.com/efbar/more-serverless/slack-message/slackmessage"
)

func Handle(w http.ResponseWriter, r *http.Request) {

	slackmessage.Serve(w, r)

}
