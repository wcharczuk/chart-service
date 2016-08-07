package request

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util"
)

const (
	HTTPREQUEST_LOG_LEVEL_ERRORS    = 1
	HTTPREQUEST_LOG_LEVEL_VERBOSE   = 2
	HTTPREQUEST_LOG_LEVEL_DEBUG     = 3
	HTTPREQUEST_LOG_LEVEL_OVER_9000 = 9001
)

//--------------------------------------------------------------------------------
// HttpResponseMeta
//--------------------------------------------------------------------------------

func newHttpResponseMeta(res *http.Response) *HttpResponseMeta {
	meta := &HttpResponseMeta{}

	if res == nil {
		return meta
	}

	meta.StatusCode = res.StatusCode
	meta.ContentLength = res.ContentLength

	content_type_header := res.Header["Content-Type"]
	if content_type_header != nil && len(content_type_header) > 0 {
		meta.ContentType = strings.Join(content_type_header, ";")
	}

	content_encoding_header := res.Header["Content-Encoding"]
	if content_encoding_header != nil && len(content_encoding_header) > 0 {
		meta.ContentEncoding = strings.Join(content_encoding_header, ";")
	}

	meta.Headers = res.Header
	return meta
}

type HttpResponseMeta struct {
	StatusCode      int
	ContentLength   int64
	ContentEncoding string
	ContentType     string
	Headers         http.Header
}

type ResponseBodyHandler func([]byte) error
type CreateTransportHook func(host url.URL, transport *http.Transport)
type IncomingResponseHook func(meta *HttpResponseMeta, content []byte)
type OutgoingRequestHook func(verb string, url *url.URL)
type OutgoingRequestBodyHook func(body []byte)
type MockedResponseHook func(verb string, url *url.URL) (bool, *HttpResponseMeta, []byte, error)

//--------------------------------------------------------------------------------
// HttpRequest
//--------------------------------------------------------------------------------

type HttpRequest struct {
	Scheme            string
	Host              string
	Path              string
	QueryString       url.Values
	Header            http.Header
	PostData          url.Values
	Cookies           []*http.Cookie
	BasicAuthUsername string
	BasicAuthPassword string
	Verb              string
	ContentType       string
	Timeout           time.Duration
	TLSCertPath       string
	TLSKeyPath        string
	Body              string
	KeepAlive         bool

	Label string

	Logger   *log.Logger
	LogLevel int

	transport *http.Transport

	createTransportHook     CreateTransportHook
	incomingResponseHook    IncomingResponseHook
	outgoingRequestHook     OutgoingRequestHook
	outgoingRequestBodyHook OutgoingRequestBodyHook
	mockHook                MockedResponseHook
}

func NewRequest() *HttpRequest {
	hr := HttpRequest{}
	hr.Scheme = "http"
	hr.Verb = "GET"
	hr.KeepAlive = false
	return &hr
}

func (hr *HttpRequest) WithLabel(label string) *HttpRequest {
	hr.Label = label
	return hr
}

func (hr *HttpRequest) WithCreateTransportHook(hook CreateTransportHook) *HttpRequest {
	hr.createTransportHook = hook
	return hr
}

func (hr *HttpRequest) WithMockedResponse(hook MockedResponseHook) *HttpRequest {
	hr.mockHook = hook
	return hr
}

func (hr *HttpRequest) WithIncomingResponseHook(hook IncomingResponseHook) *HttpRequest {
	hr.incomingResponseHook = hook
	return hr
}

func (hr *HttpRequest) WithOutgoingRequestHook(hook OutgoingRequestHook) *HttpRequest {
	hr.outgoingRequestHook = hook
	return hr
}

func (hr *HttpRequest) WithOutgoingRequestBodyHook(hook OutgoingRequestBodyHook) *HttpRequest {
	hr.outgoingRequestBodyHook = hook
	return hr
}

func (hr *HttpRequest) WithLogging() *HttpRequest {
	hr.LogLevel = HTTPREQUEST_LOG_LEVEL_ERRORS
	hr.Logger = log.New(os.Stdout, "", 0)
	return hr
}

func (hr *HttpRequest) WithLogLevel(logLevel int) *HttpRequest {
	hr.LogLevel = logLevel
	return hr
}

func (hr *HttpRequest) WithLogger(logLevel int, logger *log.Logger) *HttpRequest {
	hr.LogLevel = logLevel
	hr.Logger = logger
	return hr
}

func (hr *HttpRequest) fatalf(logLevel int, format string, args ...interface{}) {
	if hr.Logger != nil && logLevel <= hr.LogLevel {
		prefix := getLoggingPrefix(logLevel)
		hr.Logger.Fatalf(prefix+format, args...)
	}
}

func (hr *HttpRequest) fatalln(logLevel int, args ...interface{}) {
	if hr.Logger != nil && logLevel <= hr.LogLevel {
		prefix := getLoggingPrefix(logLevel)
		message := fmt.Sprint(args...)
		full_message := fmt.Sprintf("%s%s", prefix, message)
		hr.Logger.Fatalln(full_message)
	}
}

func (hr *HttpRequest) logf(logLevel int, format string, args ...interface{}) {
	if hr.Logger != nil && logLevel <= hr.LogLevel {
		prefix := getLoggingPrefix(logLevel)
		hr.Logger.Printf(prefix+format, args...)
	}
}

func (hr *HttpRequest) logln(logLevel int, args ...interface{}) {
	if hr.Logger != nil && logLevel <= hr.LogLevel {
		prefix := getLoggingPrefix(logLevel)
		message := fmt.Sprint(args...)
		full_message := fmt.Sprintf("%s%s", prefix, message)
		hr.Logger.Println(full_message)
	}
}

func (hr *HttpRequest) WithTransport(transport *http.Transport) *HttpRequest {
	hr.transport = transport
	return hr
}

func (hr *HttpRequest) WithKeepAlives() *HttpRequest {
	hr.KeepAlive = true
	hr = hr.WithHeader("Connection", "keep-alive")
	return hr
}

func (hr *HttpRequest) WithContentType(content_type string) *HttpRequest {
	hr.ContentType = content_type
	return hr
}

func (hr *HttpRequest) WithScheme(scheme string) *HttpRequest {
	hr.Scheme = scheme
	return hr
}

func (hr *HttpRequest) WithHost(host string) *HttpRequest {
	hr.Host = host
	return hr
}

func (hr *HttpRequest) WithPath(path_pattern string, args ...interface{}) *HttpRequest {
	hr.Path = fmt.Sprintf(path_pattern, args...)
	return hr
}

func (hr *HttpRequest) WithCombinedPath(components ...string) *HttpRequest {
	hr.Path = util.CombinePathComponents(components...)
	return hr
}

func (hr *HttpRequest) WithUrl(url_string string) *HttpRequest {
	working_url, _ := url.Parse(url_string)
	hr.Scheme = working_url.Scheme
	hr.Host = working_url.Host
	hr.Path = working_url.Path
	params := strings.Split(working_url.RawQuery, "&")
	hr.QueryString = url.Values{}
	var key_value []string
	for _, param := range params {
		if param != "" {
			key_value = strings.Split(param, "=")
			hr.QueryString.Set(key_value[0], key_value[1])
		}
	}
	return hr
}

func (hr *HttpRequest) WithHeader(field string, value string) *HttpRequest {
	if hr.Header == nil {
		hr.Header = http.Header{}
	}
	hr.Header.Set(field, value)
	return hr
}

func (hr *HttpRequest) WithQueryString(field string, value string) *HttpRequest {
	if hr.QueryString == nil {
		hr.QueryString = url.Values{}
	}
	hr.QueryString.Add(field, value)
	return hr
}

func (hr *HttpRequest) WithCookie(cookie *http.Cookie) *HttpRequest {
	if hr.Cookies == nil {
		hr.Cookies = []*http.Cookie{}
	}
	hr.Cookies = append(hr.Cookies, cookie)
	return hr
}

func (hr *HttpRequest) WithPostData(field string, value string) *HttpRequest {
	if hr.PostData == nil {
		hr.PostData = url.Values{}
	}
	hr.PostData.Add(field, value)
	return hr
}

func (hr *HttpRequest) WithPostDataFromObject(object interface{}) *HttpRequest {
	postDatums := util.DecomposeToPostDataAsJson(object)

	for _, item := range postDatums {
		hr.WithPostData(item.Key, item.Value)
	}

	return hr
}

func (hr *HttpRequest) WithBasicAuth(username, password string) *HttpRequest {
	hr.BasicAuthUsername = username
	hr.BasicAuthPassword = password
	return hr
}

func (hr *HttpRequest) WithTimeout(timeout time.Duration) *HttpRequest {
	hr.Timeout = timeout
	return hr
}

func (hr *HttpRequest) WithTLSCert(cert_path string) *HttpRequest {
	hr.TLSCertPath = cert_path
	return hr
}

func (hr *HttpRequest) WithTLSKey(key_path string) *HttpRequest {
	hr.TLSKeyPath = key_path
	return hr
}

func (hr *HttpRequest) WithVerb(verb string) *HttpRequest {
	hr.Verb = verb
	return hr
}

func (hr *HttpRequest) AsGet() *HttpRequest {
	hr.Verb = "GET"
	return hr
}
func (hr *HttpRequest) AsPost() *HttpRequest {
	hr.Verb = "POST"
	return hr
}
func (hr *HttpRequest) AsPut() *HttpRequest {
	hr.Verb = "PUT"
	return hr
}
func (hr *HttpRequest) AsPatch() *HttpRequest {
	hr.Verb = "PATCH"
	return hr
}
func (hr *HttpRequest) AsDelete() *HttpRequest {
	hr.Verb = "DELETE"
	return hr
}

func (hr *HttpRequest) WithJsonBody(object interface{}) *HttpRequest {
	return hr.WithBody(object, serializeJson).WithContentType("application/json")
}

func (hr *HttpRequest) WithXmlBody(object interface{}) *HttpRequest {
	return hr.WithBody(object, serializeXml).WithContentType("application/xml")
}

func (hr *HttpRequest) WithBody(object interface{}, serialize func(interface{}) string) *HttpRequest {
	return hr.WithRawBody(serialize(object))
}

func (hr *HttpRequest) WithRawBody(body string) *HttpRequest {
	hr.Body = body
	return hr
}

func (hr *HttpRequest) createUrl() url.URL {
	working_url := url.URL{Scheme: hr.Scheme, Host: hr.Host, Path: hr.Path}
	working_url.RawQuery = hr.QueryString.Encode()
	return working_url
}

func (hr *HttpRequest) RequestBody() string {
	if hr.Body != "" {
		return hr.Body
	} else if hr.PostData != nil {
		return hr.PostData.Encode()
	} else {
		return util.EMPTY
	}
}

func (hr *HttpRequest) CreateHttpRequest() (*http.Request, error) {
	working_url := hr.createUrl()

	if hr.Body != "" && hr.PostData != nil && len(hr.PostData) > 0 {
		return nil, exception.New("Cant set both a body and have post data!")
	}

	var req *http.Request
	if hr.Body != "" {
		body_req, body_req_err := http.NewRequest(hr.Verb, working_url.String(), bytes.NewBufferString(hr.Body))
		if body_req_err != nil {
			return nil, exception.Wrap(body_req_err)
		}
		req = body_req
	} else {
		if hr.PostData != nil {
			post_req, post_req_error := http.NewRequest(hr.Verb, working_url.String(), bytes.NewBufferString(hr.PostData.Encode()))
			if post_req_error != nil {
				return nil, exception.Wrap(post_req_error)
			}
			req = post_req
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		} else {
			empty_req, empty_req_err := http.NewRequest(hr.Verb, working_url.String(), nil)
			if empty_req_err != nil {
				return nil, exception.Wrap(empty_req_err)
			}
			req = empty_req
		}
	}

	if !isEmpty(hr.BasicAuthUsername) {
		req.SetBasicAuth(hr.BasicAuthUsername, hr.BasicAuthPassword)
	}

	if !isEmpty(hr.ContentType) {
		req.Header.Set("Content-Type", hr.ContentType)
	}

	for key, values := range hr.Header {
		for _, value := range values {
			req.Header.Set(key, value)
		}
	}

	if hr.Cookies != nil {
		for i := 0; i < len(hr.Cookies); i++ {
			cookie := hr.Cookies[i]
			req.AddCookie(cookie)
		}
	}

	return req, nil
}

func (hr *HttpRequest) FetchRawResponse() (*http.Response, error) {
	req, req_err := hr.CreateHttpRequest()
	if req_err != nil {
		return nil, req_err
	}

	if hr.mockHook != nil {
		did_mock_response, mocked_meta, mocked_response, mocked_response_err := hr.mockHook(hr.Verb, req.URL)
		if did_mock_response {
			buff := bytes.NewBuffer(mocked_response)
			res := http.Response{}
			buff_len := buff.Len()
			res.Body = ioutil.NopCloser(buff)
			res.ContentLength = int64(buff_len)
			res.Header = mocked_meta.Headers
			res.StatusCode = mocked_meta.StatusCode
			return &res, exception.Wrap(mocked_response_err)
		}
	}

	client := &http.Client{}
	if hr.requiresCustomTransport() {
		transport, transport_error := hr.getHttpTransport()
		if transport_error != nil {
			return nil, exception.Wrap(transport_error)
		}
		client.Transport = transport
	}

	if hr.Timeout != time.Duration(0) {
		client.Timeout = hr.Timeout
	}

	hr.logRequest(req.URL)

	res, resErr := client.Do(req)
	return res, exception.Wrap(resErr)
}

func (hr *HttpRequest) Execute() error {
	res, err := hr.FetchRawResponse()
	if res != nil && res.Body != nil {
		closeErr := res.Body.Close()
		if closeErr != nil {
			return exception.WrapMany(exception.Wrap(err), exception.Wrap(closeErr))
		}
	}
	return exception.Wrap(err)
}

func (hr *HttpRequest) ExecuteWithMeta() (*HttpResponseMeta, error) {
	res, err := hr.FetchRawResponse()
	if res != nil && res.Body != nil {
		closeErr := res.Body.Close()
		if closeErr != nil {
			return nil, exception.WrapMany(exception.Wrap(err), exception.Wrap(closeErr))
		}
	}
	meta := newHttpResponseMeta(res)
	return meta, exception.Wrap(err)
}

func (hr *HttpRequest) FetchString() (string, error) {
	response_string, _, err := hr.FetchStringWithMeta()
	return response_string, exception.Wrap(err)
}

func (hr *HttpRequest) FetchStringWithMeta() (string, *HttpResponseMeta, error) {
	res, err := hr.FetchRawResponse()
	meta := newHttpResponseMeta(res)
	if err != nil {
		return util.EMPTY, meta, exception.Wrap(err)
	}
	defer res.Body.Close()

	bytes, read_err := ioutil.ReadAll(res.Body)
	if read_err != nil {
		return util.EMPTY, meta, exception.Wrap(read_err)
	}

	meta.ContentLength = int64(len(bytes))
	hr.logResponse(meta, bytes)
	return string(bytes), meta, nil
}

func (hr *HttpRequest) FetchJsonToObject(destination interface{}) error {
	_, err := hr.deserialize(newJsonHandler(destination))
	return err
}

func (hr *HttpRequest) FetchJsonToObjectWithMeta(destination interface{}) (*HttpResponseMeta, error) {
	return hr.deserialize(newJsonHandler(destination))
}

func (hr *HttpRequest) FetchJsonToObjectWithErrorHandler(successObject interface{}, errorObject interface{}) (*HttpResponseMeta, error) {
	return hr.deserializeWithErrorHandler(newJsonHandler(successObject), newJsonHandler(errorObject))
}

func (hr *HttpRequest) FetchJsonError(errorObject interface{}) (*HttpResponseMeta, error) {
	return hr.deserializeWithErrorHandler(nil, newJsonHandler(errorObject))
}

func (hr *HttpRequest) FetchXmlToObject(destination interface{}) error {
	_, err := hr.deserialize(newXmlHandler(destination))
	return err
}

func (hr *HttpRequest) FetchXmlToObjectWithMeta(destination interface{}) (*HttpResponseMeta, error) {
	return hr.deserialize(newXmlHandler(destination))
}

func (hr *HttpRequest) FetchXmlToObjectWithErrorHandler(successObject interface{}, error_object interface{}) (*HttpResponseMeta, error) {
	return hr.deserializeWithErrorHandler(newXmlHandler(successObject), newXmlHandler(error_object))
}

func (hr *HttpRequest) FetchObjectWithSerializer(serializer ResponseBodyHandler) (*HttpResponseMeta, error) {
	meta, response_err := hr.deserialize(func(body []byte) error {
		return serializer(body)
	})
	return meta, response_err
}

func (hr *HttpRequest) requiresCustomTransport() bool {
	return (!isEmpty(hr.TLSCertPath) && !isEmpty(hr.TLSKeyPath)) || hr.transport != nil || hr.createTransportHook != nil
}

func (hr *HttpRequest) getHttpTransport() (*http.Transport, error) {
	if hr.transport != nil {
		hr.logf(HTTPREQUEST_LOG_LEVEL_DEBUG, "Service Request ==> Using Provided Transport\n")
		return hr.transport, nil
	} else {
		return hr.createHttpTransport()
	}
}

func (hr *HttpRequest) createHttpTransport() (*http.Transport, error) {
	hr.logf(HTTPREQUEST_LOG_LEVEL_DEBUG, "Service Request ==> Creating Custom Transport\n")
	transport := &http.Transport{
		DisableCompression: false,
		DisableKeepAlives:  !hr.KeepAlive,
	}

	dialer := &net.Dialer{}
	if hr.Timeout != time.Duration(0) {
		dialer.Timeout = hr.Timeout
	}
	if hr.KeepAlive {
		hr.logf(HTTPREQUEST_LOG_LEVEL_DEBUG, "Service Request ==> Transport Enabled For `keep-alive` %v\n", 30*time.Second)
		dialer.KeepAlive = 30 * time.Second
	}

	loggedDialer := func(network, address string) (net.Conn, error) {
		hr.logf(HTTPREQUEST_LOG_LEVEL_DEBUG, "Service Request ==> Transport Is Dialing %s\n", address)
		return dialer.Dial(network, address)
	}
	transport.Dial = loggedDialer

	if !isEmpty(hr.TLSCertPath) && !isEmpty(hr.TLSKeyPath) {
		if cert, err := tls.LoadX509KeyPair(hr.TLSCertPath, hr.TLSKeyPath); err != nil {
			return nil, exception.Wrap(err)
		} else {
			tlsConfig := &tls.Config{
				Certificates: []tls.Certificate{cert},
			}
			transport.TLSClientConfig = tlsConfig
		}
	}

	if hr.createTransportHook != nil {
		hr.createTransportHook(hr.createUrl(), transport)
	}

	return transport, nil
}

func (hr *HttpRequest) deserialize(handler ResponseBodyHandler) (*HttpResponseMeta, error) {
	res, err := hr.FetchRawResponse()
	meta := newHttpResponseMeta(res)

	if err != nil {
		return meta, exception.Wrap(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return meta, exception.Wrap(err)
	}

	meta.ContentLength = int64(len(body))
	hr.logResponse(meta, body)
	if handler != nil {
		err = handler(body)
	}
	return meta, exception.Wrap(err)
}

func (hr *HttpRequest) deserializeWithErrorHandler(okHandler ResponseBodyHandler, errorHandler ResponseBodyHandler) (*HttpResponseMeta, error) {
	res, err := hr.FetchRawResponse()
	meta := newHttpResponseMeta(res)

	if err != nil {
		return meta, exception.Wrap(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return meta, exception.Wrap(err)
	}

	meta.ContentLength = int64(len(body))
	hr.logResponse(meta, body)
	if res.StatusCode == http.StatusOK {
		if okHandler != nil {
			err = okHandler(body)
		}
	} else if errorHandler != nil {
		err = errorHandler(body)
	}
	return meta, exception.Wrap(err)
}

func (hr *HttpRequest) logRequest(url *url.URL) {
	if hr.outgoingRequestHook != nil {
		hr.outgoingRequestHook(hr.Verb, url)
	}
	if hr.outgoingRequestBodyHook != nil {
		hr.outgoingRequestBodyHook([]byte(hr.RequestBody()))
	}
	hr.logf(HTTPREQUEST_LOG_LEVEL_VERBOSE, "Service Request ==> %s %s\n", hr.Verb, url.String())
}

func (hr *HttpRequest) logResponse(meta *HttpResponseMeta, responseBody []byte) {
	if hr.incomingResponseHook != nil {
		hr.incomingResponseHook(meta, responseBody)
	}
	hr.logf(HTTPREQUEST_LOG_LEVEL_VERBOSE, "Service Response ==> %s", string(responseBody))
}

//--------------------------------------------------------------------------------
// Unexported Utility Functions
//--------------------------------------------------------------------------------

func newJsonHandler(object interface{}) ResponseBodyHandler {
	return func(body []byte) error {
		return deserializeJson(object, string(body))
	}
}

func newXmlHandler(object interface{}) ResponseBodyHandler {
	return func(body []byte) error {
		return deserializeXml(object, string(body))
	}
}

func deserializeJson(object interface{}, body string) error {
	decoder := json.NewDecoder(bytes.NewBufferString(body))
	decodeErr := decoder.Decode(object)
	return exception.Wrap(decodeErr)
}

func deserializeJsonFromReader(object interface{}, body io.Reader) error {
	decoder := json.NewDecoder(body)
	decodeErr := decoder.Decode(object)
	return exception.Wrap(decodeErr)
}

func deserializePostBody(object interface{}, body io.ReadCloser) error {
	defer body.Close()
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return exception.Wrap(err)
	}

	return deserializeJson(object, string(bodyBytes))
}

func serializeJson(object interface{}) string {
	b, _ := json.Marshal(object)
	return string(b)
}

func serializeJsonToReader(object interface{}) io.Reader {
	b, _ := json.Marshal(object)
	return bytes.NewBufferString(string(b))
}

func deserializeXml(object interface{}, body string) error {
	return deserializeXmlFromReader(object, bytes.NewBufferString(body))
}

func deserializeXmlFromReader(object interface{}, reader io.Reader) error {
	decoder := xml.NewDecoder(reader)
	return decoder.Decode(object)
}

func serializeXml(object interface{}) string {
	b, _ := xml.Marshal(object)
	return string(b)
}

func serializeXmlToReader(object interface{}) io.Reader {
	b, _ := xml.Marshal(object)
	return bytes.NewBufferString(string(b))
}

func getLoggingPrefix(logLevel int) string {
	return fmt.Sprintf("HttpRequest (%s): ", formatLogLevel(logLevel))
}

func formatLogLevel(logLevel int) string {
	switch logLevel {
	case HTTPREQUEST_LOG_LEVEL_ERRORS:
		return "ERRORS"
	case HTTPREQUEST_LOG_LEVEL_VERBOSE:
		return "VERBOSE"
	case HTTPREQUEST_LOG_LEVEL_DEBUG:
		return "DEBUG"
	default:
		return "UNKNOWN"
	}
}

func isEmpty(str string) bool {
	return len(str) == 0
}
