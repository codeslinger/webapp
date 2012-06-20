// vim:set ts=2 sw=2 et ai ft=go:
package webapp

import (
  "net/http"
  "time"
)

// --- COOKIE CONSTRUCTORS --------------------------------------------------

// Create a new session cookie. It will expire at the end of the browser's
// current session.
func NewSessionCookie(name, value string) *http.Cookie {
  c := &http.Cookie {
    Name: name,
    Value: value,
    MaxAge: 0,
  }
  return c
}

// Create a new persistent cookie.
func NewPersistentCookie(name, value string) *http.Cookie {
  c := &http.Cookie {
    Name: name,
    Value: value,
    Expires: time.Unix(2147483647, 0),
  }
  return c
}

// Create a record that will cause the given cookie to be deleted when sent
// to the browser.
func DeleteCookie(name string) *http.Cookie {
  c := &http.Cookie {
    Name: name,
    Value: "DELETED",
    MaxAge: -1,
  }
  return c
}

