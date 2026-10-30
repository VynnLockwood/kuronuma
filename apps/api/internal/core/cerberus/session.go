package cerberus

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// randString returns n random bytes encoded as unpadded base64url.
func randString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// newSessionToken returns a fresh opaque token (given to the browser) and its
// SHA-256 hash (stored in the database).
func newSessionToken() (token string, hash []byte, err error) {
	token, err = randString(32)
	if err != nil {
		return "", nil, err
	}
	return token, hashToken(token), nil
}

func hashToken(token string) []byte {
	h := sha256.Sum256([]byte(token))
	return h[:]
}
