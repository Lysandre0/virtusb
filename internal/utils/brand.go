package utils

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"sync"
)

var (
	vidPidCache = map[string][2]string{
		"sandisk":  {"0781", "5567"},
		"kingston": {"0951", "1666"},
		"corsair":  {"1b1c", "1a0a"},
		"samsung":  {"04e8", "61b6"},
		"generic":  {"13fe", "4200"},
	}
	cacheMutex sync.RWMutex
)

func VidPid(brand string) (string, string) {
	brandLower := strings.ToLower(brand)

	cacheMutex.RLock()
	if result, exists := vidPidCache[brandLower]; exists {
		cacheMutex.RUnlock()
		return result[0], result[1]
	}
	cacheMutex.RUnlock()

	return "13fe", "4200"
}

func SerialFor(brand string) string {
	brandLower := strings.ToLower(brand)

	if brandLower == "sandisk" {
		return "4C530001" + randHex(12)
	}
	return randHex(16)
}

func randHex(n int) string {
	bufferSize := (n + 1) / 2
	b := make([]byte, bufferSize)

	_, _ = rand.Read(b)

	s := strings.ToUpper(hex.EncodeToString(b))

	if len(s) > n {
		s = s[:n]
	}

	return s
}
