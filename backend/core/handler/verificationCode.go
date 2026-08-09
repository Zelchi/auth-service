package handler

import (
	"crypto/sha256"
	"encoding/hex"
)

func digestVerificationCode(code string) string {
	digest := sha256.Sum256([]byte(code))
	return hex.EncodeToString(digest[:])
}
