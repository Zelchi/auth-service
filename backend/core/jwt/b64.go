package jwt

import "encoding/base64"

func b64(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}
