package plutus

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// maxMinor is the largest amount parseMinor can return (10 whole digits plus 99 hundredths).
const maxMinor = 999_999_999_999

// parseMinor converts a plain decimal amount such as "450", "450.5" or "1234.56"
// into hundredths of the currency unit. It rejects signs, exponents and more than
// two decimal places so that a misread number is reported rather than rounded.
func parseMinor(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("amount is required")
	}
	whole, frac, hasDot := strings.Cut(s, ".")
	if whole == "" || len(whole) > 10 || !allDigits(whole) {
		return 0, errors.New("amount must be a plain positive number like 450 or 450.50")
	}
	if hasDot && (frac == "" || len(frac) > 2 || !allDigits(frac)) {
		return 0, errors.New("amount can have at most 2 decimal places")
	}
	for len(frac) < 2 {
		frac += "0"
	}
	w, _ := strconv.ParseInt(whole, 10, 64)
	f, _ := strconv.ParseInt(frac, 10, 64)
	m := w*100 + f
	if m <= 0 {
		return 0, errors.New("amount must be greater than 0")
	}
	return m, nil
}

func formatMinor(m int64) string { return fmt.Sprintf("%d.%02d", m/100, m%100) }

func allDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}
