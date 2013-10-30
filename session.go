// vim:set ts=2 sw=2 et ai ft=go:
package webapp

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultSessionDuration = 14 * 86400
)

var (
	SessionKeyUndefined = errors.New("SessionKey not defined")
	InvalidSignature    = errors.New("Signature is invalid")
	SessionExpired      = errors.New("Session has expired")
)

type Session struct {
	data map[string]interface{}
	ttl  int64
}

// --- SESSION API ----------------------------------------------------------

// Get a value from the session by key.
func (s *Session) Get(key string) interface{} {
	return s.data[key]
}

// Set a value in the session by key.
func (s *Session) Set(key string, val interface{}) {
	s.data[key] = val
}

// Delete a value from the session.
func (s *Session) Delete(key string) {
	delete(s.data, key)
}

// --- SESSION INTERNALS -----------------------------------------------------

// Private constructor for Session records.
func newSession(ttl int64) *Session {
	return &Session{
		data: make(map[string]interface{}),
		ttl:  ttl,
	}
}

// Deserialize and validate a marshalled Session record.
func (s *Session) unmarshal(raw, key string) error {
	if len(key) == 0 {
		return SessionKeyUndefined
	}
	pieces := strings.SplitN(raw, "|", 3)
	data := pieces[0]
	timestamp := pieces[1]
	signature := pieces[2]
	if s.sign(key, []byte(data), timestamp) != signature {
		return InvalidSignature
	}
	t, _ := strconv.ParseInt(timestamp, 0, 64)
	if time.Now().Unix()-s.ttl > t {
		return SessionExpired
	}
	return s.decode(data)
}

// Sign and serialize a Session record for use in a Set-Cookie HTTP header.
func (s *Session) marshal(key string) (string, error) {
	if len(key) == 0 {
		return "", SessionKeyUndefined
	}
	encoded, err := s.encode()
	if err != nil {
		return "", err
	}
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature := s.sign(key, encoded, timestamp)
	final := strings.Join([]string{string(encoded), timestamp, signature}, "|")
	return final, nil
}

// Generate a signature for a session.
func (s *Session) sign(key string, data []byte, timestamp string) string {
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write(data)
	mac.Write([]byte(timestamp))
	return fmt.Sprintf("%02x", mac.Sum(nil))
}

// Encode session data into format suitable for HTTP.
func (s *Session) encode() ([]byte, error) {
	data, err := json.Marshal(s.data)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	encoder.Write(data)
	encoder.Close()
	return buf.Bytes(), nil
}

// Decode session data from HTTP format.
func (s *Session) decode(data string) error {
	buf := bytes.NewBufferString(data)
	decoder := base64.NewDecoder(base64.StdEncoding, buf)
	ioutil.ReadAll(decoder)
	if err := json.Unmarshal(buf.Bytes(), s.data); err != nil {
		return err
	}
	return nil
}
