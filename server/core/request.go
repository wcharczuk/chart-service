package core

import (
	"fmt"
	"time"

	"github.com/blendlabs/go-request"
	"github.com/blendlabs/go-util"
)

// NewRequest returns a request with some extra helpers bolted on.
func NewRequest() *request.HTTPRequest {
	return request.NewHTTPRequest().OnRequest(func(req *request.HTTPRequestMeta) {
		fmt.Printf("%s Outgoing %s %s\n",
			util.Color(time.Now().UTC().Format(time.RFC3339), util.ColorGray),
			util.Color(req.Verb, util.ColorBlue),
			req.URL.String(),
		)
	})
}
