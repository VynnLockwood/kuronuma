package cerberus

import (
	"math"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var allScopes = []Scope{
	{Name: "plutus:read", Description: "Read your spending records and summaries"},
	{Name: "plutus:write", Description: "Add and delete your spending records"},
	{Name: "categories:read", Description: "Read your category list"},
	{Name: "categories:write", Description: "Create, rename and merge your categories"},
	{Name: "atlas:read", Description: "Read your tasks, to-dos and notes"},
	{Name: "atlas:write", Description: "Add, change and delete your tasks, to-dos and notes"},
}

func (h *oh) consentPage(vals url.Values, extra url.Values, hdr map[string]string) (*http.Response, string) {
	h.t.Helper()
	all := url.Values{}
	for k, v := range vals {
		all[k] = v
	}
	for k, v := range extra {
		all[k] = v
	}
	rec := h.send("GET", "/api/oauth/authorize?"+all.Encode(), "", "", []*http.Cookie{h.session}, hdr)
	if rec.Code != http.StatusOK {
		h.t.Fatalf("consent status = %d body=%s", rec.Code, rec.Body.String())
	}
	return rec.Result(), rec.Body.String()
}

func mustContain(t *testing.T, page string, wants ...string) {
	t.Helper()
	for _, w := range wants {
		if !strings.Contains(page, w) {
			t.Errorf("page lacks %q", w)
		}
	}
}

func mustNotContain(t *testing.T, page string, nots ...string) {
	t.Helper()
	for _, n := range nots {
		if strings.Contains(page, n) {
			t.Errorf("page must not contain %q", n)
		}
	}
}

// The consent page has to answer "should I approve this?": a question as the heading, a plain checklist, the
// permissions grouped by area in everyday words, and an access level.
func TestConsentPageExplainsTheChoice(t *testing.T) {
	h := newOHScopes(t, quietLog(), allScopes)
	cid := h.registerClient("Google", testRedirect)
	res, page := h.consentPage(authzValues(cid, testRedirect, "st", ""), nil, nil)

	mustContain(t, page,
		`<html lang="en">`,
		"Connect Google to your Kuronuma?",
		"Check before you approve",
		"Approve if you just added Kuronuma as a connector in Gemini, Claude or ChatGPT.",
		"Deny if you didn’t start this, or someone sent you the link.", // typographic apostrophe: a plain one is HTML-escaped
		"Spending", "See your spending records and totals", "Add and delete spending records, and create categories as it saves",
		"Categories", "See your category list", "Create, rename and merge categories",
		"Tasks, to-dos and notes", "See your tasks, to-dos and notes", "Add, change and delete tasks, to-dos and notes, and create categories as it saves",
		`name="access" value="full" checked`, `name="access" value="read"`, "Read and change", "Read only",
		`<img src="/brand/kuronuma-logo.png"`,
		`<title>Connect to Kuronuma</title>`, `<span>KURONUMA</span>`,
		"http://localhost:7777", // where the user lands afterwards
	)
	mustNotContain(t, page, "Only approve if you started this connection yourself", "Nemo", "NEMO") // [keep-nemosyne]

	// Cloudflare rewrites e-mail addresses in HTML into a link that needs a script this page cannot run (its
	// security policy forbids scripts); the markers keep the address as plain text.
	if !regexp.MustCompile(`<!--email_off-->[^<@]+@[^<]+<!--/email_off-->`).MatchString(page) {
		t.Errorf("the signed-in address must be wrapped in email_off markers")
	}
	csp := res.Header.Get("Content-Security-Policy")
	mustContain(t, csp, "default-src 'none'", "img-src 'self'", "frame-ancestors 'none'")
	mustNotContain(t, csp, "script-src", "form-action", "'unsafe-eval'")
}

func TestConsentPageShowsOnlyWhatWasRequested(t *testing.T) {
	h := newOHScopes(t, quietLog(), allScopes)
	cid := h.registerClient("App", testRedirect)

	_, page := h.consentPage(authzValues(cid, testRedirect, "st", "plutus:read"), nil, nil)
	mustContain(t, page, "Spending", "See your spending records and totals", `name="access" value="read"`)
	mustNotContain(t, page, "Categories", "Tasks, to-dos and notes", "Add and delete spending records")

	// A request that only asks to change things cannot be narrowed to read-only, so the choice is not offered.
	_, page = h.consentPage(authzValues(cid, testRedirect, "st", "plutus:write"), nil, nil)
	mustContain(t, page, "Add and delete spending records")
	mustNotContain(t, page, `value="read"`, "Read only")
}

func TestConsentPageInThai(t *testing.T) {
	h := newOHScopes(t, quietLog(), allScopes)
	cid := h.registerClient("Google", testRedirect)
	vals := authzValues(cid, testRedirect, "st", "")

	_, th := h.consentPage(vals, url.Values{"lang": {"th"}}, nil)
	mustContain(t, th,
		`<html lang="th">`, "เชื่อมต่อ Google กับคุโรนุมะของคุณไหม?", "<title>เชื่อมต่อกับคุโรนุมะ</title>", "ตรวจสอบก่อนอนุมัติ",
		"ค่าใช้จ่าย", "หมวดหมู่", "งาน สิ่งที่ต้องทำ และโน้ต",
		"อ่านและแก้ไขได้", "อ่านอย่างเดียว", "อนุมัติ", "ปฏิเสธ",
		`name="lang" value="th"`, "lang=en", // the form remembers the language, and there is a link back
	)
	mustNotContain(t, th, "Connect Google to your Kuronuma?", "Check before you approve", "Nemo") // [keep-nemosyne]

	thAccept := map[string]string{"Accept-Language": "th-TH,th;q=0.9,en;q=0.8"}
	_, page := h.consentPage(vals, nil, thAccept)
	mustContain(t, page, `<html lang="th">`)
	_, page = h.consentPage(vals, url.Values{"lang": {"en"}}, thAccept)
	mustContain(t, page, `<html lang="en">`, "Connect Google to your Kuronuma?", "lang=th")
	_, page = h.consentPage(vals, nil, map[string]string{"Accept-Language": "en-US,en;q=0.9"})
	mustContain(t, page, `<html lang="en">`)
	_, page = h.consentPage(vals, url.Values{"lang": {"xx"}}, nil)
	mustContain(t, page, `<html lang="en">`)

	// The language link keeps the rest of the request, so switching does not lose the connection attempt.
	_, page = h.consentPage(vals, nil, nil)
	m := regexp.MustCompile(`href="(/api/oauth/authorize\?[^"]+)"`).FindStringSubmatch(page)
	if m == nil {
		t.Fatalf("no language link on the page")
	}
	link, _ := url.Parse(strings.ReplaceAll(m[1], "&amp;", "&"))
	if link.Query().Get("lang") != "th" || link.Query().Get("client_id") != cid || link.Query().Get("state") != "st" {
		t.Errorf("language link = %s", m[1])
	}
}

// The access level decides which scopes the authorization code carries. The server derives them from the
// original request, never from a list in the form.
func TestApprovalAccessLevel(t *testing.T) {
	grant := func(t *testing.T, scope string, extra url.Values) ([]string, int) {
		t.Helper()
		h := newOHScopes(t, quietLog(), allScopes)
		cid := h.registerClient("App", testRedirect)
		vals := authzValues(cid, testRedirect, "st", scope)
		csrf, _ := h.consent(vals)
		form := url.Values{}
		for k, v := range vals {
			form[k] = v
		}
		for k, v := range extra {
			form[k] = v
		}
		form.Set("csrf", csrf)
		form.Set("decision", "approve")
		rec := h.send("POST", "/api/oauth/authorize", "application/x-www-form-urlencoded", form.Encode(),
			[]*http.Cookie{h.session}, map[string]string{"Origin": testOrigin})
		var scopes []string
		for _, c := range h.store.codes {
			scopes = c.Scopes
		}
		return scopes, rec.Code
	}
	both := "plutus:read plutus:write atlas:read atlas:write"
	cases := []struct {
		name      string
		scope     string
		extra     url.Values
		want      []string
		wantCode  int
		wantNoMsg bool
	}{
		{"full access", both, url.Values{"access": {"full"}}, []string{"plutus:read", "plutus:write", "atlas:read", "atlas:write"}, http.StatusSeeOther, false},
		{"no access field means full", both, nil, []string{"plutus:read", "plutus:write", "atlas:read", "atlas:write"}, http.StatusSeeOther, false},
		{"read only keeps only the read scopes", both, url.Values{"access": {"read"}}, []string{"plutus:read", "atlas:read"}, http.StatusSeeOther, false},
		{"read only when everything was asked for", "", url.Values{"access": {"read"}}, []string{"plutus:read", "categories:read", "atlas:read"}, http.StatusSeeOther, false},
		{"unknown access level is refused", both, url.Values{"access": {"admin"}}, nil, http.StatusBadRequest, true},
		{"read only is impossible for a write-only request", "plutus:write", url.Values{"access": {"read"}}, nil, http.StatusBadRequest, true},
		{"a scope list in the form changes nothing", "plutus:read", url.Values{"access": {"full"}, "granted": {"plutus:write"}, "scopes": {"plutus:write"}}, []string{"plutus:read"}, http.StatusSeeOther, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, code := grant(t, c.scope, c.extra)
			if code != c.wantCode {
				t.Fatalf("status = %d, want %d", code, c.wantCode)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("scopes = %v, want %v", got, c.want)
			}
		})
	}
}

func TestReadOnlyTokenCarriesOnlyReadScopes(t *testing.T) {
	h := newOHScopes(t, quietLog(), allScopes)
	cid := h.registerClient("App", testRedirect)
	vals := authzValues(cid, testRedirect, "st", "plutus:read plutus:write")
	csrf, _ := h.consent(vals)
	form := url.Values{"csrf": {csrf}, "decision": {"approve"}, "access": {"read"}}
	for k, v := range vals {
		form[k] = v
	}
	rec := h.send("POST", "/api/oauth/authorize", "application/x-www-form-urlencoded", form.Encode(),
		[]*http.Cookie{h.session}, map[string]string{"Origin": testOrigin})
	loc, _ := url.Parse(rec.Header().Get("Location"))
	tr := decodeToken(t, h.token(codeForm(cid, loc.Query().Get("code"), testRedirect, testVerifier)))
	if tr.Scope != "plutus:read" {
		t.Fatalf("token scope = %q, want plutus:read", tr.Scope)
	}
}

func TestErrorPageIsBranded(t *testing.T) {
	h := newOH(t)
	rec := h.send("GET", "/api/oauth/authorize?client_id=kuro_c_nope&redirect_uri=http://x.example/<b>", "", "", nil, nil)
	body := rec.Body.String()
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	mustContain(t, body, "KURONUMA", "unknown client_id", "/brand/kuronuma-logo.png")
	mustNotContain(t, body, "NEMO", "Nemo") // [keep-nemosyne]
	mustNotContain(t, body, "<b>")
}

// The pages are decorated with two small inline SVG motifs (consent) and one (error). They are hidden from
// assistive technology, and the page still runs no script, under an unchanged security policy.
func TestConsentPageHasMotifsAndNoScripts(t *testing.T) {
	h := newOHScopes(t, quietLog(), allScopes)
	cid := h.registerClient("Google", testRedirect)
	res, page := h.consentPage(authzValues(cid, testRedirect, "st", ""), nil, nil)

	if n := strings.Count(page, "<svg"); n != 2 {
		t.Errorf("consent page has %d <svg> elements, want 2", n)
	}
	if n := strings.Count(page, `<svg aria-hidden="true"`); n != 2 {
		t.Errorf("%d of the svg elements are aria-hidden, want 2", n)
	}
	mustContain(t, page, `focusable="false"`)
	mustNotContain(t, page, "<script", "onclick", "onload", "javascript:")
	if got, want := res.Header.Get("Content-Security-Policy"), "default-src 'none'; style-src 'unsafe-inline'; img-src 'self'; frame-ancestors 'none'"; got != want {
		t.Errorf("CSP = %q, want %q", got, want)
	}
}

func TestErrorPageHasAMotif(t *testing.T) {
	h := newOH(t)
	rec := h.send("GET", "/api/oauth/authorize?client_id=kuro_c_nope&redirect_uri=http://x.example/", "", "", nil, nil)
	body := rec.Body.String()
	if n := strings.Count(body, "<svg"); n != 1 {
		t.Errorf("error page has %d <svg> elements, want 1", n)
	}
	mustContain(t, body, `<svg aria-hidden="true"`)
	mustContain(t, body, `data-motif="plane"`)
	mustNotContain(t, body, "M24 4 C25.5") // the sparkle's path: the error page uses the paper plane now
	mustNotContain(t, body, "<script")
}

func luminance(hex string) float64 {
	v, err := strconv.ParseUint(strings.TrimPrefix(hex, "#"), 16, 32)
	if err != nil || len(hex) != 7 {
		panic("bad colour " + hex)
	}
	ch := func(x uint64) float64 {
		c := float64(x) / 255
		if c <= 0.03928 {
			return c / 12.92
		}
		return math.Pow((c+0.055)/1.055, 2.4)
	}
	return 0.2126*ch(v>>16&255) + 0.7152*ch(v>>8&255) + 0.0722*ch(v&255)
}

func contrast(a, b string) float64 {
	la, lb := luminance(a), luminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

func TestPageColoursAreAccessible(t *testing.T) {
	for _, c := range []struct{ got, want string }{
		{colBackground, "#fffafb"}, {colInk, "#2a2024"}, {colMuted, "#7a5c68"}, {colPrimary, "#c23b69"},
		{colLink, "#b03361"}, {colBorder, "#efd3dd"}, {colSurface, "#fdf0f4"},
	} {
		if c.got != c.want {
			t.Errorf("colour constant %q, want %q", c.got, c.want)
		}
	}
	for _, text := range []string{colInk, colMuted, colLink} {
		for _, bg := range []string{colBackground, colSurface} {
			if c := contrast(text, bg); c < 4.5 {
				t.Errorf("%s on %s has contrast %.2f, want at least 4.5", text, bg, c)
			}
		}
	}
	if c := contrast("#ffffff", colPrimary); c < 4.5 {
		t.Errorf("white on %s has contrast %.2f, want at least 4.5", colPrimary, c)
	}
	// The page is built from the constants, not from a second copy of the values.
	mustContain(t, pageStyle, colBackground, colInk, colMuted, colPrimary, colLink, colBorder, colSurface)
	mustNotContain(t, pageStyle, "gradient")
}
