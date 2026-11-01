// Package webguard holds the checks for requests that the website itself sends to the API with the visitor's
// session cookie (the dashboard). The cookie is SameSite=Lax, and these checks close the rest of the door on
// cross-site request forgery: a page on another site cannot add a custom header without a CORS preflight (and no
// CORS headers are served), and an HTML form cannot send JSON.
package webguard

import (
	"errors"
	"mime"
	"net/http"
	"net/url"
	"strings"
)

// Header must be present, with the value "1", on every request from our own pages.
const Header = "X-Kuronuma-Web"

// One error per rule, so a test (or a log line) can say which one refused the request.
var (
	ErrHeader      = errors.New("webguard: the " + Header + " header is missing")
	ErrContentType = errors.New("webguard: the content type must be application/json")
	ErrOrigin      = errors.New("webguard: the Origin is not this site")
	ErrFetchSite   = errors.New("webguard: the request did not come from this site's own page")
	ErrConfig      = errors.New("webguard: the public URL is not a usable origin")
)

// Check returns nil when r looks like a request sent by our own page, and the first rule it breaks otherwise.
// publicURL is the site's public address; if it is not usable every request fails closed.
func Check(r *http.Request, publicURL string) error {
	want, err := origin(publicURL)
	if err != nil {
		return err
	}
	if r.Header.Get(Header) != "1" {
		return ErrHeader
	}
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != "application/json" {
		return ErrContentType
	}
	if o := r.Header.Get("Origin"); o != "" && strings.ToLower(o) != want {
		return ErrOrigin
	}
	// Browsers send Sec-Fetch-Site on every request. "same-site" (a sibling subdomain) is refused as well as
	// "cross-site": only our own origin is our page. Clients that do not send it (not browsers) are not affected.
	if v := r.Header.Get("Sec-Fetch-Site"); v != "" && v != "same-origin" && v != "none" {
		return ErrFetchSite
	}
	return nil
}

// origin reduces a public URL to scheme://host[:port] in lower case.
func origin(publicURL string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(publicURL))
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") {
		return "", ErrConfig
	}
	return strings.ToLower(u.Scheme + "://" + u.Host), nil
}
