package web

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// New returns a new app.
func New() *App {
	return &App{
		router:             httprouter.New(),
		name:               "Web",
		staticRewriteRules: map[string][]*RewriteRule{},
		staticHeaders:      map[string]http.Header{},
		requestStartHandlers: []RequestEventHandler{func(r *RequestContext) {
			r.onRequestStart()
		}},
		requestCompleteHandlers: []RequestEventHandler{func(r *RequestContext) {
			r.onRequestEnd()
		}},
		requestErrorHandlers: []RequestEventErrorHandler{func(r *RequestContext, err interface{}) {
			if r != nil && r.logger != nil {
				r.logger.Error(err)
			}
		}},
	}
}

// AppStartDelegate is a function that is run on start. Typically you use this to initialize the app.
type AppStartDelegate func(app *App) error

// App is the server for the app.
type App struct {
	name string

	logger    Logger
	router    *httprouter.Router
	viewCache *template.Template

	apiResultProvider  *APIResultProvider
	viewResultProvider *ViewResultProvider

	startDelegate AppStartDelegate

	requestStartHandlers    []RequestEventHandler
	requestCompleteHandlers []RequestEventHandler
	requestErrorHandlers    []RequestEventErrorHandler

	requestLogFormat string

	staticRewriteRules map[string][]*RewriteRule
	staticHeaders      map[string]http.Header

	tx *sql.Tx

	port string
}

// Name returns the app name.``
func (a *App) Name() string {
	return a.name
}

// SetName sets the app name
func (a *App) SetName(name string) {
	a.name = name
}

// RequestLogFormat returns a custom w3c request log format if set.
func (a *App) RequestLogFormat() string {
	return a.requestLogFormat
}

// SetRequestLogFormat sets a custom w3c request log format.
func (a *App) SetRequestLogFormat(format string) {
	a.requestLogFormat = format
}

// Logger returns the logger for the app.
func (a *App) Logger() Logger {
	return a.logger
}

// SetLogger sets the logger.
func (a *App) SetLogger(l Logger) {
	a.logger = l
}

// ViewCache gets the view cache for the app.
func (a *App) ViewCache() *template.Template {
	return a.viewCache
}

// SetViewCache sets the view cache for the app.
func (a *App) SetViewCache(viewCache *template.Template) {
	a.viewCache = viewCache
}

// IsolateTo sets the app to use a transaction for *all* requests.
// Caveat: only use during testing.
func (a *App) IsolateTo(tx *sql.Tx) {
	a.tx = tx
}

// Port returns the port for the app.
func (a *App) Port() string {
	if len(a.port) != 0 {
		return a.port
	}
	envVar := os.Getenv("PORT")
	if len(envVar) != 0 {
		return envVar
	}

	return "8080"
}

// SetPort sets the port the app listens on.
func (a *App) SetPort(port string) {
	a.port = port
}

// OnStart lets you register a task that is run before the server starts.
// Typically this delegate sets up the database connection and other init items.
func (a *App) OnStart(action AppStartDelegate) {
	a.startDelegate = action
}

// Start starts the server and binds to the given address.
func (a *App) Start() error {
	bindAddr := fmt.Sprintf(":%s", a.Port())
	server := &http.Server{
		Addr:    bindAddr,
		Handler: a,
	}
	return a.StartWithServer(server)
}

// StartWithServer starts the app on a custom server.
// This lets you configure things like TLS keys and
// other options.
func (a *App) StartWithServer(server *http.Server) error {
	if a.startDelegate != nil {
		a.Logf("%s Startup tasks starting", a.Name())
		err := a.startDelegate(a)
		if err != nil {
			a.Errorf("%s Startup tasks error: %v", a.Name(), err)
			return err
		}
		a.Logf("%s Startup tasks complete", a.Name())
	}

	// this is the only property we will set of the server
	// i.e. the server handler (which is this app)
	server.Handler = a
	a.Logf("%s Started, listening on %s", a.Name(), server.Addr)

	return server.ListenAndServe()
}

// Register registers a controller with the app's router.
func (a *App) Register(c Controller) {
	c.Register(a)
}

// InitViewCache caches templates by path.
func (a *App) InitViewCache(paths ...string) error {
	views, err := template.ParseFiles(paths...)
	if err != nil {
		return err
	}
	a.viewCache = template.Must(views, nil)
	return nil
}

// GET registers a GET request handler.
func (a *App) GET(path string, action ControllerAction, middleware ...ControllerMiddleware) {
	a.router.GET(path, a.renderAction(a.nestMiddleware(action, middleware...)))
}

// OPTIONS registers a OPTIONS request handler.
func (a *App) OPTIONS(path string, action ControllerAction, middleware ...ControllerMiddleware) {
	a.router.OPTIONS(path, a.renderAction(a.nestMiddleware(action, middleware...)))
}

// HEAD registers a HEAD request handler.
func (a *App) HEAD(path string, action ControllerAction, middleware ...ControllerMiddleware) {
	a.router.HEAD(path, a.renderAction(a.nestMiddleware(action, middleware...)))
}

// PUT registers a PUT request handler.
func (a *App) PUT(path string, action ControllerAction, middleware ...ControllerMiddleware) {
	a.router.PUT(path, a.renderAction(a.nestMiddleware(action, middleware...)))
}

// PATCH registers a PATCH request handler.
func (a *App) PATCH(path string, action ControllerAction, middleware ...ControllerMiddleware) {
	a.router.PATCH(path, a.renderAction(a.nestMiddleware(action, middleware...)))
}

// POST registers a POST request actions.
func (a *App) POST(path string, action ControllerAction, middleware ...ControllerMiddleware) {
	a.router.POST(path, a.renderAction(a.nestMiddleware(action, middleware...)))
}

// DELETE registers a DELETE request handler.
func (a *App) DELETE(path string, action ControllerAction, middleware ...ControllerMiddleware) {
	a.router.DELETE(path, a.renderAction(a.nestMiddleware(action, middleware...)))
}

// --------------------------------------------------------------------------------
// Logging methods
// --------------------------------------------------------------------------------

// Log logs a message to the logger if one is provisioned.
func (a *App) Log(args ...interface{}) {
	if a.logger != nil {
		a.logger.Log(args...)
	}
}

// Logf logs a message to the logger if one is provisioned.
func (a *App) Logf(format string, args ...interface{}) {
	if a.logger != nil {
		a.logger.Logf(format, args...)
	}
}

// Error logs a message to the logger if one is provisioned.
func (a *App) Error(args ...interface{}) {
	if a.logger != nil {
		a.logger.Error(args...)
	}
}

// Errorf logs a message to the logger if one is provisioned.
func (a *App) Errorf(format string, args ...interface{}) {
	if a.logger != nil {
		a.logger.Errorf(format, args...)
	}
}

// --------------------------------------------------------------------------------
// Static Result Methods
// --------------------------------------------------------------------------------

// StaticRewrite adds a rewrite rule for a specific statically served path.
// Make sure to serve the static path with (app).Static(path, root).
func (a *App) StaticRewrite(path, match string, action RewriteAction) error {
	expr, err := regexp.Compile(match)
	if err != nil {
		return err
	}
	a.staticRewriteRules[path] = append(a.staticRewriteRules[path], &RewriteRule{
		MatchExpression: match,
		expr:            expr,
		Action:          action,
	})

	return nil
}

// StaticHeader adds a header for the given static path.
func (a *App) StaticHeader(path, key, value string) {
	if _, hasHeaders := a.staticHeaders[path]; !hasHeaders {
		a.staticHeaders[path] = http.Header{}
	}
	a.staticHeaders[path].Add(key, value)
}

// Static serves files from the given file system root.
// The path must end with "/*filepath", files are then served from the local
// path /defined/root/dir/*filepath.
// For example if root is "/etc" and *filepath is "passwd", the local file
// "/etc/passwd" would be served.
// Internally a http.FileServer is used, therefore http.NotFound is used instead
// of the Router's NotFound handler.
// To use the operating system's file system implementation,
// use http.Dir:
//     router.ServeFiles("/src/*filepath", http.Dir("/var/www"))
func (a *App) Static(path string, root http.FileSystem) {
	if len(path) < 10 || path[len(path)-10:] != "/*filepath" {
		panic("path must end with /*filepath in path '" + path + "'")
	}

	a.router.GET(path, a.renderAction(a.staticAction(path, root)))
}

// staticAction returns a ControllerAction for a given static path and root.
func (a *App) staticAction(path string, root http.FileSystem) ControllerAction {
	fileServer := http.FileServer(root)

	return func(r *RequestContext) ControllerResult {

		var staticRewriteRules []*RewriteRule
		var staticHeaders http.Header

		if rules, hasRules := a.staticRewriteRules[path]; hasRules {
			staticRewriteRules = rules
		}

		if headers, hasHeaders := a.staticHeaders[path]; hasHeaders {
			staticHeaders = headers
		}

		filePath, _ := r.RouteParameter("filepath")

		return &StaticResult{
			FilePath:     filePath,
			FileServer:   fileServer,
			RewriteRules: staticRewriteRules,
			Headers:      staticHeaders,
		}
	}
}

// --------------------------------------------------------------------------------
// Router internal methods
// --------------------------------------------------------------------------------

// SetNotFoundHandler sets the not found handler.
func (a *App) SetNotFoundHandler(handler ControllerAction) {
	a.router.NotFound = newHandleShim(a, handler)
}

// SetMethodNotAllowedHandler sets the not found handler.
func (a *App) SetMethodNotAllowedHandler(handler ControllerAction) {
	a.router.MethodNotAllowed = newHandleShim(a, handler)
}

// SetPanicHandler sets the not found handler.
func (a *App) SetPanicHandler(handler PanicControllerAction) {
	a.router.PanicHandler = func(w http.ResponseWriter, r *http.Request, err interface{}) {
		a.renderAction(func(r *RequestContext) ControllerResult {
			a.onRequestError(r, err)
			return handler(r, err)
		})(w, r, httprouter.Params{})
	}
}

func (a *App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	a.router.ServeHTTP(w, req)
}

// --------------------------------------------------------------------------------
// Events
// --------------------------------------------------------------------------------

// OnRequestStart triggers the onRequestStart event.
func (a *App) onRequestStart(r *RequestContext) {
	if len(a.requestStartHandlers) > 0 {
		for _, handler := range a.requestStartHandlers {
			handler(r)
		}
	}
}

// OnRequestComplete triggers the onRequestStart event.
func (a *App) onRequestComplete(r *RequestContext) {
	if len(a.requestCompleteHandlers) > 0 {
		for _, handler := range a.requestCompleteHandlers {
			handler(r)
		}
	}
}

// OnRequestError triggers the onRequestStart event.
func (a *App) onRequestError(r *RequestContext, err interface{}) {
	if len(a.requestErrorHandlers) > 0 {
		for _, handler := range a.requestErrorHandlers {
			handler(r, err)
		}
	}
}

// RequestStartHandler fires before an action handler is run.
func (a *App) RequestStartHandler(handler RequestEventHandler) {
	a.requestStartHandlers = append(a.requestStartHandlers, handler)
}

// RequestCompleteHandler fires after an action handler is run.
func (a *App) RequestCompleteHandler(handler RequestEventHandler) {
	a.requestCompleteHandlers = append(a.requestCompleteHandlers, handler)
}

// RequestErrorHandler fires if there is an error logged.
func (a *App) RequestErrorHandler(handler RequestEventErrorHandler) {
	a.requestErrorHandlers = append(a.requestErrorHandlers, handler)
}

// --------------------------------------------------------------------------------
// Testing Methods
// --------------------------------------------------------------------------------

// Mock returns a request bulider to facilitate mocking requests.
func (a *App) Mock() *MockRequestBuilder {
	return NewMockRequestBuilder(a)
}

// --------------------------------------------------------------------------------
// Render Methods
// --------------------------------------------------------------------------------

// RequestContext creates an http context.
func (a *App) requestContext(w ResponseWriter, r *http.Request, p RouteParameters) *RequestContext {
	hc := NewRequestContext(w, r, p)
	if len(a.requestLogFormat) != 0 {
		hc.requestLogFormat = a.requestLogFormat
	}
	hc.tx = a.tx
	hc.logger = a.logger
	hc.api = NewAPIResultProvider(a, hc)
	hc.view = NewViewResultProvider(a, hc)
	return hc
}

// nestMiddleware reads the middleware variadic args and organizes the calls recursively in the order they appear.
func (a *App) nestMiddleware(action ControllerAction, middleware ...ControllerMiddleware) ControllerAction {
	if len(middleware) == 0 {
		return action
	}

	var nest = func(a, b ControllerMiddleware) ControllerMiddleware {
		if b == nil {
			return a
		}
		return func(action ControllerAction) ControllerAction {
			return a(b(action))
		}
	}

	var metaAction ControllerMiddleware
	for _, step := range middleware {
		metaAction = nest(step, metaAction)
	}
	return metaAction(action)
}

// renderAction is the translation step from ControllerAction to httprouter.Handle.
func (a *App) renderAction(action ControllerAction) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		a.commonResponseHeaders(w)

		var response ResponseWriter
		if a.shouldCompressOutput(r) {
			w.Header().Set("Content-Encoding", "gzip")
			response = NewCompressedResponseWriter(w)
		} else {
			response = NewRawResponseWriter(w)
		}

		context := a.pipelineInit(response, r, NewRouteParameters(p))
		a.renderResult(action, context)
		a.pipelineComplete(context)
	}
}

func (a *App) shouldCompressOutput(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
}

func (a *App) commonResponseHeaders(w http.ResponseWriter) {
	w.Header().Set("Vary", "Accept-Encoding")
	w.Header().Set("X-Served-By", "github.com/wcharczuk/go-web")
	w.Header().Set("X-Frame-Options", "SAMEORIGIN")
	w.Header().Set("X-Xss-Protection", "1; mode=block")
	w.Header().Set("X-Content-Type-Options", "nosniff")
}

func (a *App) pipelineInit(w ResponseWriter, r *http.Request, p RouteParameters) *RequestContext {
	context := a.requestContext(w, r, p)
	a.onRequestStart(context)
	return context
}

func (a *App) renderResult(action ControllerAction, context *RequestContext) {
	result := action(context)
	if result != nil {
		err := result.Render(context.Response, context.Request)
		if err != nil {
			if a.logger != nil {
				a.logger.Error(err)
			}
		}
	}
}

func (a *App) pipelineComplete(context *RequestContext) {
	context.Response.Flush()
	context.setStatusCode(context.Response.StatusCode())
	context.setContentLength(context.Response.ContentLength())
	a.onRequestComplete(context)
	context.LogRequest()
}
