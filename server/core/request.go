package core

import (
	logger "github.com/blendlabs/go-logger"
	"github.com/blendlabs/go-request"
)

// NewRequest returns a request with some extra helpers bolted on.
func NewRequest() *request.HTTPRequest {
	return request.NewHTTPRequest().WithLogger(logger.Default())
}
