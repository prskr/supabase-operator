package supabase

import (
	"crypto/rand"
	"encoding/hex"
)

func RandomJWTSecret() ([]byte, error) {
	jwtSecretBytes := make([]byte, ServiceConfig.JWT.Defaults.SecretLength)

	if _, err := rand.Read(jwtSecretBytes); err != nil {
		return nil, err
	}

	jwtSecretHex := make([]byte, hex.EncodedLen(len(jwtSecretBytes)))
	hex.Encode(jwtSecretHex, jwtSecretBytes)

	return jwtSecretHex, nil
}
