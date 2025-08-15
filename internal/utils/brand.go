package utils

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

// VidPid returns VID/PID identifiers for a given brand
func VidPid(brand string) (string, string) {
	switch strings.ToLower(brand) {
	case "sandisk":
		return "0781", "5567"
	case "kingston":
		return "0951", "1666"
	case "corsair":
		return "1b1c", "1a0a"
	case "samsung":
		return "04e8", "61b6"
	default:
		return "13fe", "4200" // generic Phison
	}
}

// SerialFor generates a serial number for a given brand
func SerialFor(brand string) string {
	if strings.ToLower(brand) == "sandisk" {
		return "4C530001" + randHex(12)
	}
	return randHex(16)
}

// randHex generates a random hexadecimal string
func randHex(n int) string {
	b := make([]byte, (n+1)/2)
	_, _ = rand.Read(b)
	s := strings.ToUpper(hex.EncodeToString(b))
	if len(s) > n {
		s = s[:n]
	}
	return s
}
