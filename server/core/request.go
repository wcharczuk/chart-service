package core

import (
	"fmt"
	"net/http"
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
	}).OnResponse(func(meta *request.HTTPResponseMeta, content []byte) {
		statusText := util.Color(fmt.Sprintf("%d", meta.StatusCode), util.ColorGreen)
		if meta.StatusCode >= http.StatusInternalServerError {
			statusText = util.Color(fmt.Sprintf("%d", meta.StatusCode), util.ColorRed)
		} else if meta.StatusCode > http.StatusBadRequest {
			statusText = util.Color(fmt.Sprintf("%d", meta.StatusCode), util.ColorYellow)
		}
		fmt.Printf("%s Outgoing Response %s\n",
			util.Color(time.Now().UTC().Format(time.RFC3339), util.ColorGray),
			statusText,
		)

	}) /*.OnResponse(func(meta *request.HTTPResponseMeta, content []byte) {
		statusText := util.Color(fmt.Sprintf("%d", meta.StatusCode), util.ColorGreen)
		if meta.StatusCode >= http.StatusInternalServerError {
			statusText = util.Color(fmt.Sprintf("%d", meta.StatusCode), util.ColorRed)
		} else if meta.StatusCode > http.StatusBadRequest {
			statusText = util.Color(fmt.Sprintf("%d", meta.StatusCode), util.ColorYellow)
		}

		fmt.Printf("%s Response %s %s\n",
			util.Color(time.Now().UTC().Format(time.RFC3339), util.ColorGray),
			statusText,
			string(content),
		)
	})*/
}
