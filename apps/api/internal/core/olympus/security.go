// Package olympus serves the local admin page: a read-only view of the configuration and a button
// to import a backup. It is meant for the owner's own machine only; see guard for how the page
// defends itself against other websites open in the same browser.
package olympus

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// guard protects a page that listens on localhost. Being on 127.0.0.1 does not stop a website the
// owner has open from sending requests to it (cross-site form posts) or from pointing its own
// hostname at 127.0.0.1 (DNS rebinding), so every request is checked:
//
//   - Host must be localhost or 127.0.0.1 on the admin port, otherwise 404 (stops DNS rebinding).
//   - Only GET, HEAD and POST are accepted.
//   - POST must come from this page: Origin (or Referer when Origin is absent) must match, and the
//     form must carry this process's random token (stops cross-site posts).
//   - Every response carries strict security headers and no cookies are ever set.
type guard struct {
	hosts   map[string]bool // allowed Host header values, lower case
	origins map[string]bool // allowed Origin values, lower case
	token   string
	maxBody int64
}

// newGuard builds a guard for the listener at addr (host:port; only the port is used). maxBody caps
// the size of a POST body, framing included.
func newGuard(addr string, maxBody int64) (*guard, error) {
	_, port, err := net.SplitHostPort(addr)
	if err != nil || port == "" {
		return nil, fmt.Errorf("admin address %q must look like host:port", addr)
	}
	if maxBody <= 0 {
		return nil, errors.New("maximum body size must be positive")
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	g := &guard{
		hosts:   map[string]bool{},
		origins: map[string]bool{},
		token:   base64.RawURLEncoding.EncodeToString(raw),
		maxBody: maxBody,
	}
	for _, h := range []string{"localhost", "127.0.0.1"} {
		hp := h + ":" + port
		g.hosts[hp] = true
		g.origins["http://"+hp] = true
	}
	return g, nil
}

// Token is the per-process CSRF token the import form must send in its "csrf" field.
func (g *guard) Token() string { return g.token }

// Wrap applies the checks above in front of next. For an accepted POST the multipart form has
// already been parsed fully in memory (maxBody bounds it), so next can call r.FormFile and nothing
// is written to disk.
func (g *guard) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setSecurityHeaders(w.Header())

		if !g.hosts[strings.ToLower(r.Host)] {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		switch r.Method {
		case http.MethodGet, http.MethodHead:
		case http.MethodPost:
			if !g.sameOrigin(r) {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, g.maxBody)
			if err := r.ParseMultipartForm(g.maxBody); err != nil {
				var tooBig *http.MaxBytesError
				if errors.As(err, &tooBig) {
					w.WriteHeader(http.StatusRequestEntityTooLarge)
				} else {
					w.WriteHeader(http.StatusBadRequest)
				}
				return
			}
			if subtle.ConstantTimeCompare([]byte(r.PostFormValue("csrf")), []byte(g.token)) != 1 {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		default:
			w.Header().Set("Allow", "GET, HEAD, POST")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// sameOrigin reports whether a POST was sent by this page. Origin decides when present; Referer is
// only the fallback for browsers that omit Origin. A request with neither is refused.
func (g *guard) sameOrigin(r *http.Request) bool {
	if o := r.Header.Get("Origin"); o != "" {
		return g.origins[strings.ToLower(o)]
	}
	ref := r.Header.Get("Referer")
	if ref == "" {
		return false
	}
	u, err := url.Parse(ref)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	return g.origins[strings.ToLower(u.Scheme+"://"+u.Host)]
}

func setSecurityHeaders(h http.Header) {
	h.Set("Cache-Control", "no-store")
	h.Set("Content-Security-Policy", "default-src 'none'; style-src 'self'; form-action 'self'")
	h.Set("X-Frame-Options", "DENY")
	h.Set("Referrer-Policy", "same-origin") // not no-referrer: that makes browsers send "Origin: null" on form posts
	h.Set("X-Content-Type-Options", "nosniff")
}
