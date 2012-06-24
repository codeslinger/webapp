// vim:set ts=2 sw=2 et ai ft=go:
package webapp

import (
  "net/http"
  "strconv"
  "time"
)

const (
  persistentCookieAge = 2147483647
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
  session       *Session
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

// Retrieve a cookie from this request by name. Error is not nil if no cookie
// with the given name could be found.
func (req *Request) GetCookie(name string) (*http.Cookie, error) {
  return req.r.Cookie(name)
}

// Set a cookie on the client browser. Expires indicates how many seconds in
// the future the cookie is to expire. (use -1 for no expiry)
func (req *Request) SetCookie(cookie *http.Cookie) {
  http.SetCookie(req.w, cookie)
}

// Delete a cookie from the client browser.
func (req *Request) DeleteCookie(name string) {
  http.SetCookie(req.w, &http.Cookie{Name: name, MaxAge: -1})
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
  if req.session != nil {
    sval, err := req.session.marshal(req.app.SessionKey)
    if err != nil {
      req.app.Log.Error("failed to write session: %v", err)
    } else if len(sval) > 0 {
      http.SetCookie(req.w, &http.Cookie{
        Name:     req.app.SessionName,
        Value:    sval,
        MaxAge:   persistentCookieAge,
        HttpOnly: true,
      })
    }
  }
  req.replied = true
  req.w.WriteHeader(req.status)
  if req.r.Method != "HEAD" && req.contentLength > 0 {
    req.w.Write([]byte(body))
  }
}

// Retrieve the session record.
func (req *Request) Session() *Session {
  if req.session == nil {
    if err := req.loadSession(); err != nil {
      req.app.Log.Error("failed to load session: %v", err)
    }
  }
  return req.session
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

// Load session from cookies, creating a blank session if no session cookie
// exists. Returns error if we failed to get the session cookie or it did not
// validate. (i.e. was tampered with)
func (req *Request) loadSession() error {
  req.session = newSession(req.app.SessionDuration)
  cookie, err := req.GetCookie(req.app.SessionName)
  if err == http.ErrNoCookie {
    return nil
  } else if err != nil {
    return err
  }
  return req.session.unmarshal(cookie.Value, req.app.SessionKey)
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

