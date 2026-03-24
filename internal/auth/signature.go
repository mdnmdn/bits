package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// GenerateHMACSHA256Base64 returns a base64 encoded HMAC-SHA256 signature.
func GenerateHMACSHA256Base64(message, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// GenerateHMACSHA256Hex returns a hex encoded HMAC-SHA256 signature.
func GenerateHMACSHA256Hex(message, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}
