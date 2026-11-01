package webguard

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const site = "https://kuronuma.example"

// req builds a request that passes every rule; mutate changes it.
func req(mutate func(r *http.Request)) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/api/web/call", strings.NewReader("{}"))
	r.Header.Set(Header, "1")
	r.Header.Set("Content-Type", "application/json")
	if mutate != nil {
		mutate(r)
	}
	return r
}

func TestCheck(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(r *http.Request)
		url    string
		want   error // nil means the request passes
	}{
		{"all rules met", nil, site, nil},
		{"same-origin Origin and Fetch Metadata", func(r *http.Request) {
			r.Header.Set("Origin", site)
			r.Header.Set("Sec-Fetch-Site", "same-origin")
		}, site, nil},
		{"public URL with a trailing slash", func(r *http.Request) { r.Header.Set("Origin", site) }, site + "/", nil},
		{"charset after the media type", func(r *http.Request) { r.Header.Set("Content-Type", "application/json; charset=utf-8") }, site, nil},
		{"Sec-Fetch-Site none", func(r *http.Request) { r.Header.Set("Sec-Fetch-Site", "none") }, site, nil},
		{"no Origin and no Fetch Metadata (non-browser client)", nil, site, nil},

		{"header missing", func(r *http.Request) { r.Header.Del(Header) }, site, ErrHeader},
		{"header has another value", func(r *http.Request) { r.Header.Set(Header, "true") }, site, ErrHeader},
		{"header empty", func(r *http.Request) { r.Header.Set(Header, "") }, site, ErrHeader},

		{"content type missing", func(r *http.Request) { r.Header.Del("Content-Type") }, site, ErrContentType},
		{"text/plain (a form can send it)", func(r *http.Request) { r.Header.Set("Content-Type", "text/plain") }, site, ErrContentType},
		{"urlencoded form", func(r *http.Request) { r.Header.Set("Content-Type", "application/x-www-form-urlencoded") }, site, ErrContentType},
		{"multipart form", func(r *http.Request) { r.Header.Set("Content-Type", "multipart/form-data; boundary=x") }, site, ErrContentType},
		{"almost json", func(r *http.Request) { r.Header.Set("Content-Type", "application/jsonp") }, site, ErrContentType},

		{"another origin", func(r *http.Request) { r.Header.Set("Origin", "https://evil.example") }, site, ErrOrigin},
		{"http instead of https", func(r *http.Request) { r.Header.Set("Origin", "http://kuronuma.example") }, site, ErrOrigin},
		{"another port", func(r *http.Request) { r.Header.Set("Origin", "https://kuronuma.example:8443") }, site, ErrOrigin},
		{"a subdomain", func(r *http.Request) { r.Header.Set("Origin", "https://a.kuronuma.example") }, site, ErrOrigin},
		{"Origin null", func(r *http.Request) { r.Header.Set("Origin", "null") }, site, ErrOrigin},
		{"Origin with a path", func(r *http.Request) { r.Header.Set("Origin", site+"/x") }, site, ErrOrigin},

		{"cross-site", func(r *http.Request) { r.Header.Set("Sec-Fetch-Site", "cross-site") }, site, ErrFetchSite},
		{"same-site (a sibling subdomain is not our page)", func(r *http.Request) { r.Header.Set("Sec-Fetch-Site", "same-site") }, site, ErrFetchSite},

		{"empty public URL fails closed", nil, "", ErrConfig},
		{"public URL without a host fails closed", nil, "https://", ErrConfig},
		{"public URL that is not http(s) fails closed", nil, "ftp://kuronuma.example", ErrConfig},
		{"garbage public URL fails closed", nil, "::::", ErrConfig},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := Check(req(tc.mutate), tc.url)
			if !errors.Is(err, tc.want) || (tc.want == nil && err != nil) {
				t.Fatalf("Check = %v, want %v", err, tc.want)
			}
		})
	}
}

// A cross-site HTML form can only send these three content types and cannot add a custom header, so the plain
// form POST that is the classic CSRF is refused on the header alone.
func TestACrossSiteFormPostIsRefused(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/api/web/call", strings.NewReader("tool=delete_transaction"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Set("Origin", "https://evil.example")
	r.Header.Set("Sec-Fetch-Site", "cross-site")
	if err := Check(r, site); err == nil {
		t.Fatal("a cross-site form post must be refused")
	}
}
