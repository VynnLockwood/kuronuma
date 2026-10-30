package cerberus

import (
	"testing"
	"time"
)

func TestSafeNext(t *testing.T) {
	cases := map[string]string{
		"":                       "/",
		"/":                      "/",
		"/dashboard":             "/dashboard",
		"/a/b?x=1&y=2":           "/a/b?x=1&y=2",
		"//evil.example":         "/",
		"https://evil.example":   "/",
		"http://evil.example/x":  "/",
		"/\\evil.example":        "/",
		"/\t/evil.example":       "/", // browsers strip the tab, leaving //evil.example
		"/\n/evil.example":       "/",
		"/\r/evil.example":       "/",
		"relative/path":          "/",
		"javascript:alert(1)":    "/",
		"/%2F/evil.example":      "/%2F/evil.example", // stays on our host
		"/ok#fragment":           "/ok#fragment",
		"/with space is escaped": "/with space is escaped",
	}
	for in, want := range cases {
		if got := safeNext(in); got != want {
			t.Errorf("safeNext(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFlowRoundTripAndExpiry(t *testing.T) {
	now := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	in := flowState{State: "s", Nonce: "n", Verifier: "v", Next: "/x", Expires: now.Add(time.Minute).Unix()}

	sealed, err := sealFlow(testSecret, in)
	if err != nil {
		t.Fatal(err)
	}
	out, err := openFlow(testSecret, sealed, now)
	if err != nil || out != in {
		t.Fatalf("round trip: got %+v err=%v", out, err)
	}
	if _, err := openFlow(testSecret, sealed, now.Add(2*time.Minute)); err == nil {
		t.Fatal("expired flow must be rejected")
	}
	if _, err := openFlow(testSecret, "garbage", now); err == nil {
		t.Fatal("garbage must be rejected")
	}
	if _, err := openFlow(testSecret, sealed+"A", now); err == nil {
		t.Fatal("modified signature must be rejected")
	}
}

func TestSessionTokens(t *testing.T) {
	a, ha, err := newSessionToken()
	if err != nil {
		t.Fatal(err)
	}
	b, hb, _ := newSessionToken()
	if a == b {
		t.Fatal("tokens must be unique")
	}
	if len(a) < 43 {
		t.Fatalf("token too short: %d chars", len(a))
	}
	if string(ha) == string(hb) || string(ha) == a {
		t.Fatal("hashes must differ from each other and from the token")
	}
	if string(hashToken(a)) != string(ha) {
		t.Fatal("hash must be reproducible from the token")
	}
}
