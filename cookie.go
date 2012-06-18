// vim:set ts=2 sw=2 et ai ft=go:
package webapp

import (
  "bytes"
  "errors"
  "fmt"
  "time"
)

const (
  emptyString      = ""
  persistentCookie = -1
  sessionCookie    = -2
  deletedCookie    = -3
)

type Cookie struct {
  name     string
  value    string
  expires  int64
  path     string
  domain   string
  secure   bool
  httpOnly bool
}

// --- COOKIE CONSTRUCTORS --------------------------------------------------

// Create a new cookie with that will expire `expires` seconds into the
// future.
func NewCookie(name, value string, expires int64) (*Cookie, error) {
  if expires < 0 {
    return nil, errors.New("cookie expiry cannot be negative")
  }
  return createCookie(name, value, expires), nil
}

// Create a new session cookie. It will expire at the end of the browser's
// current session.
func NewSessionCookie(name, value string) *Cookie {
  return createCookie(name, value, sessionCookie)
}

// Create a new persistent cookie.
func NewPersistentCookie(name, value string) *Cookie {
  return createCookie(name, value, persistentCookie)
}

// Create a record that will cause the given cookie to be deleted when sent
// to the browser.
func DeleteCookie(name string) *Cookie {
  return createCookie(name, "DELETED", deletedCookie)
}

// --- COOKIE API ------------------------------------------------------------

func (c *Cookie) GetPath() string {
  return c.path
}

func (c *Cookie) SetPath(path string) {
  c.path = path
}

func (c *Cookie) GetDomain() string {
  return c.domain
}

func (c *Cookie) SetDomain(domain string) {
  c.domain = domain
}

func (c *Cookie) IsSecure() bool {
  return c.secure
}

func (c *Cookie) SetSecure(secure bool) {
  c.secure = secure
}

func (c *Cookie) IsHttpOnly() bool {
  return c.httpOnly
}

func (c *Cookie) SetHttpOnly(httpOnly bool) {
  c.httpOnly = httpOnly
}

// --- INTERNAL FUNCTIONS ---------------------------------------------------

// Internal constructor to build cookie record.
func createCookie(name, value string, expires int64) *Cookie {
  c := &Cookie {
    name:     name,
    value:    value,
    expires:  expires,
    path:     emptyString,
    domain:   emptyString,
    secure:   false,
    httpOnly: false,
  }
  return c
}

// Marshal a cookie for transmission from server to client.
func (c *Cookie) marshal() string {
  var buf bytes.Buffer

  fmt.Fprintf(&buf, "%s=%s", c.name, c.value)
  t := c.buildExpiry()
  if t != emptyString {
    fmt.Fprintf(&buf, t)
  }
  if c.path != emptyString {
    fmt.Fprintf(&buf, "; Path=%s", c.path)
  }
  if c.domain != emptyString {
    fmt.Fprintf(&buf, "; Domain=%s", c.domain)
  }
  if c.secure {
    fmt.Fprintf(&buf, "; Secure")
  }
  if c.httpOnly {
    fmt.Fprintf(&buf, "; HttpOnly")
  }
  return buf.String()
}

// Determine the Expires portion of the Set-Cookie header for this cookie.
func (c *Cookie) buildExpiry() string {
  var t time.Time

  switch c.expires {
  case persistentCookie:
    t = time.Unix(2147483647, 0)
  case sessionCookie:
    return emptyString
  case deletedCookie:
    t = time.Unix(0, 0)
  default:
    t = time.Unix(time.Now().Unix() + c.expires, 0)
  }
  return fmt.Sprintf("; Expires=%s", httpDate(t))
}

