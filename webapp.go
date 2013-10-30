// vim:set ts=2 sw=2 et ai ft=go:
package webapp

import (
	"bytes"
	"fmt"
	"github.com/codeslinger/log"
	"net/http"
	"regexp"
	"runtime"
	"time"
)

// --- WEBAPP API ------------------------------------------------------------

// The RouteHandler is the type a function should be if it wishes to register
// for handling a route.
//
// If a request arrives matching the pattern for a route, its RouteHandler will
// be called to respond to the request. The RouteHandler func is given a
// pointer to a Request record and a list of argument values extracted from the
// route pattern given.
//
// E.g. if a route is registered with the pattern: "/foo/(\d+)/bar/(\w+)" Then
// args will contain two values, the first being the string matched between the
// "foo" and the "bar" parts of the request URI and the second being the string
// matched between the "bar" and the end of the string.
type RouteHandler func(*Request, []string)

// A Webapp is the main edifice for a web application.
type Webapp struct {
	Log             *log.Logger
	LogHits         bool
	SessionName     string
	SessionKey      string
	SessionDuration int64
	host            string
	port            int
	routes          []route
}

// Create a new Webapp instance. The host and port on which to listen are given,
// as well as the Logger to use.
func New(host string, port int, log *log.Logger) *Webapp {
	app := &Webapp{
		Log:             log,
		LogHits:         true,
		host:            host,
		port:            port,
		SessionName:     "_session",
		SessionDuration: DefaultSessionDuration,
	}
	return app
}

// Start the Webapp listening and serving requests.
func (app *Webapp) Run() {
	addr := fmt.Sprintf("%s:%d", app.host, app.port)
	s := &http.Server{
		Addr:           addr,
		Handler:        app,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	app.Log.Info("application started: listening on %s", addr)
	err := s.ListenAndServe()
	if err != nil {
		app.Log.Error(err)
	}
}

// --- ROUTE REGISTRATION ---------------------------------------------------

// Register a route for a given pattern for GET requests. (will also be called
// for HEAD requests)
func (app *Webapp) Get(pattern string, handler RouteHandler) {
	app.registerRoute(pattern, "GET", handler)
}

// Register a route for a given pattern for POST requests.
func (app *Webapp) Post(pattern string, handler RouteHandler) {
	app.registerRoute(pattern, "POST", handler)
}

// Register a route for a given pattern for PUT requests.
func (app *Webapp) Put(pattern string, handler RouteHandler) {
	app.registerRoute(pattern, "PUT", handler)
}

// Register a route for a given pattern for DELETE requests.
func (app *Webapp) Delete(pattern string, handler RouteHandler) {
	app.registerRoute(pattern, "DELETE", handler)
}

// --- APP INTERNALS --------------------------------------------------------

// Describes an individual route pattern and its associated handler.
type route struct {
	pattern string
	re      *regexp.Regexp
	method  string
	handler RouteHandler
}

// Main callback for Webapp instance on receipt of new HTTP request.
func (app *Webapp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := newRequest(w, r, app)
	path := r.URL.Path
	for i := 0; i < len(app.routes); i++ {
		route := app.routes[i]
		if r.Method != route.method && !(r.Method == "HEAD" && route.method == "GET") {
			continue
		}
		if !route.re.MatchString(path) {
			continue
		}
		match := route.re.FindStringSubmatch(path)
		err := app.protect(route.handler, req, match[1:])
		if err != nil {
			req.Reply(500, "Internal server error")
		}
		return
	}
	req.NotFound("<h1>Not found</h1>")
	if req.app.LogHits {
		req.logHit()
	}
}

// Does the work of registering a route pattern and handler with this
// Webapp instance.
func (app *Webapp) registerRoute(pattern string, method string, handler RouteHandler) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		app.Log.Critical("could not compile route pattern: %q", pattern)
	}
	app.routes = append(app.routes, route{pattern, re, method, handler})
}

// Run a RouteHandler safely, ensuring that panics inside handlers are trapped
// and logged.
func (app *Webapp) protect(handler RouteHandler, req *Request, args []string) (e interface{}) {
	defer func() {
		if err := recover(); err != nil {
			e = err
			var buf bytes.Buffer
			fmt.Fprintf(&buf, "handler crashed: %v\n", err)
			for i := 2; ; i++ {
				_, file, line, ok := runtime.Caller(i)
				if !ok {
					break
				}
				fmt.Fprintf(&buf, "! %s:%d\n", file, line)
			}
			app.Log.Error(buf.String())
		}
	}()
	handler(req, args)
	return
}
