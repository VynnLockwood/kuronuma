package cerberus

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"time"
)

// flowState is what we remember between redirecting to Google and the callback.
// It travels in a short-lived HMAC-signed cookie so the API stays stateless here.
type flowState struct {
	State    string `json:"s"`
	Nonce    string `json:"n"`
	Verifier string `json:"v"` // PKCE code verifier
	Next     string `json:"x"` // local path to return to after sign-in
	Expires  int64  `json:"e"` // unix seconds
}

const flowDomain = "kuronuma-oauth-flow-v1:"

var errBadFlow = errors.New("cerberus: invalid or expired login flow")

func flowMAC(secret []byte, payload string) []byte {
	m := hmac.New(sha256.New, secret)
	m.Write([]byte(flowDomain))
	m.Write([]byte(payload))
	return m.Sum(nil)
}

func sealFlow(secret []byte, f flowState) (string, error) {
	raw, err := json.Marshal(f)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(raw)
	sig := base64.RawURLEncoding.EncodeToString(flowMAC(secret, payload))
	return payload + "." + sig, nil
}

func openFlow(secret []byte, value string, now time.Time) (flowState, error) {
	payload, sig, ok := strings.Cut(value, ".")
	if !ok {
		return flowState{}, errBadFlow
	}
	got, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil || !hmac.Equal(got, flowMAC(secret, payload)) {
		return flowState{}, errBadFlow
	}
	raw, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return flowState{}, errBadFlow
	}
	var f flowState
	if err := json.Unmarshal(raw, &f); err != nil {
		return flowState{}, errBadFlow
	}
	if now.Unix() > f.Expires {
		return flowState{}, errBadFlow
	}
	return f, nil
}

// safeNext only lets sign-in return to a same-site path, never another origin.
func safeNext(next string) string {
	if next == "" || len(next) > 2000 || next[0] != '/' || strings.HasPrefix(next, "//") {
		return "/"
	}
	for _, r := range next {
		// Browsers strip tabs/newlines and treat "\" like "/", so reject all of them.
		if r < 0x20 || r == 0x7f || r == '\\' {
			return "/"
		}
	}
	u, err := url.Parse(next)
	if err != nil || u.IsAbs() || u.Host != "" {
		return "/"
	}
	return next
}
