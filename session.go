// vim:set ts=2 sw=2 et ai ft=go:
package webapp

import (
  /*
  "crypto/hmac"
  "crypto/sha1"
  "encoding/base64"
  "encoding/json"
  */
  "errors"
)

var (
  SessionKeyUndefined = errors.New("SessionKey not defined")
)

type Session struct {
  data  map[string]string
  dirty bool
}

// --- SESSION API ----------------------------------------------------------

// Get a value from the session by key.
func (s *Session) Get(key string) string {
  return s.data[key]
}

// Set a value in the session by key.
func (s *Session) Set(key string, val string) {
  s.data[key] = val
  s.dirty = true
}

// Delete a value from the session.
func (s *Session) Delete(key string) {
  delete(s.data, key)
  s.dirty = true
}

// --- SESSION INTERNALS -----------------------------------------------------

// Private constructor for Session records.
func newSession() *Session {
  return &Session{
    data:  make(map[string]string),
    dirty: false,
  }
}

// Deserialize and validate a marshalled Session record.
func (s *Session) unmarshal(raw string, key string) error {
  if len(key) == 0 {
    return SessionKeyUndefined
  }
  return nil
}

// Sign and serialize a Session record for use in a Set-Cookie HTTP header.
func (s *Session) marshal(name, key string) (string, error) {
  if len(key) == 0 {
    return "", SessionKeyUndefined
  }
  // don't marshal the session if nothing changed
  if !s.dirty {
    return "", nil
  }
  return "", nil
}

