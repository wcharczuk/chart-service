package core

import "github.com/blendlabs/go-request"

// NewRequest returns a request with some extra helpers bolted on.
func NewRequest() *request.HTTPRequest {
	return request.NewHTTPRequest()
}
