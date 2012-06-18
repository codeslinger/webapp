// vim:set ts=2 sw=2 et ai ft=go:
package webapp

import (
  "strings"
  "time"
)

// Format a given time for use with HTTP protocol.
func httpDate(t time.Time) string {
  f := t.UTC().Format(time.RFC1123)
  if strings.HasSuffix(f, "UTC") {
    f = f[0:len(f)-3] + "GMT"
  }
  return f
}

