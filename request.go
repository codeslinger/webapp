// vim:set ts=2 sw=2 et ai ft=go:
package webapp

import (
  "net/http"
  "strconv"
  "time"
)

// --- REQUEST API ----------------------------------------------------------

// The Request record encapsulates all the of the state required to handle
// an HTTP request/response cycle.
type Request struct {
  w             http.ResponseWriter
  r             *http.Request
  app           *Webapp
  status        int
  contentLength int
  contentType   string
  date          time.Time
  replied       bool
}

// Sets the named header to the given value. This will override any existing
// value for the named header with the given value.
func (req *Request) SetHeader(name, val string) {
  req.w.Header().Set(name, val)
}

// Add an instance of a header to the output with the given value. This allows
// for more than one header with the given name to be output (e.g. Set-Cookie).
func (req *Request) AddHeader(name, val string) {
  req.w.Header().Add(name, val)
}

// Set a cookie on the client browser. Expires indicates how many seconds in
// the future the cookie is to expire. (use -1 for no expiry)
func (req *Request) SetCookie(cookie *http.Cookie) {
  http.SetCookie(req.w, cookie)
}

// Respond to the request with an HTTP OK (200) status code and the given
// response body. Use an empty string for no body.
func (req *Request) OK(body string) {
  req.Reply(http.StatusOK, body)
}

// Respond to the request with an HTTP Not Found (404) status and the given
// response body. Use an empty string for no body.
func (req *Request) NotFound(body string) {
  req.Reply(http.StatusNotFound, body)
}

// Respond to the request with the given status code and response body. Use
// an empty string for no body.
func (req *Request) Reply(status int, body string) {
  if req.replied {
    req.app.Log.Critical("this context has already been replied to!")
  }
  req.status = status
  req.contentLength = len(body)
  req.SetHeader("Date", httpDate(req.date))
  if req.contentLength > 0 {
    req.SetHeader("Content-Type", req.contentType)
    req.SetHeader("Content-Length", strconv.Itoa(req.contentLength))
  }
  if req.status >= 400 {
    req.SetHeader("Connection", "close")
  }
  req.replied = true
  req.w.WriteHeader(req.status)
  if req.r.Method != "HEAD" && req.contentLength > 0 {
    req.w.Write([]byte(body))
  }
}

// --- REQUEST INTERNALS ----------------------------------------------------

// Private constructor for Request records. These should only be created by
// an Webapp instance.
func newRequest(w http.ResponseWriter, r *http.Request, app *Webapp) *Request {
  req := &Request {
    w:             w,
    r:             r,
    app:           app,
    status:        200,
    contentLength: 0,
    contentType:   "text/html; charset=utf-8",
    date:          time.Now(),
    replied:       false,
  }
  return req
}

// Record pertinent request and response information in the log.
func (req *Request) logHit() {
  bytesSent := "-"
  if req.contentLength > 0 {
    bytesSent = strconv.Itoa(req.contentLength)
  }
  req.app.Log.Info("hit: %s %s %s %s %d %s\n",
                   req.r.RemoteAddr,
                   req.r.Method,
                   req.r.URL.Path,
                   req.r.Proto,
                   req.status,
                   bytesSent)
}

