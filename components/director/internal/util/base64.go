package util

import "encoding/base64"

// TryDecodeBase64 attempts to base64 decode the input multiple times until
// it is no longer base64 encoded. If the input is not base64 encoded, it
// will return it as-is.
func TryDecodeBase64(s string) []byte {
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return []byte(s)
	}

	for {
		encoded := decoded

		buffer := make([]byte, base64.StdEncoding.DecodedLen(len(encoded)))
		length, err := base64.StdEncoding.Decode(buffer, encoded)
		if err != nil {
			return decoded
		}

		decoded = buffer[:length]
	}
}
