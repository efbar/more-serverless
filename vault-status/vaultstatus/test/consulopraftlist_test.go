package testing

import (
	"testing"
)

func TestFunc(t *testing.T) {

	tt := []struct {
		aclenabled  bool
		loglevel    string
		contentType string
	}{
		{
			aclenabled:  true,
			loglevel:    "INFO",
			contentType: "text/plain",
		},
		{
			aclenabled:  true,
			loglevel:    "INFO",
			contentType: "application/json",
		},
	}

	for _, tr := range tt {

	}

}
