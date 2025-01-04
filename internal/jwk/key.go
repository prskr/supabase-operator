package jwk

import (
	"encoding/base64"
	"encoding/json"
)

type KeyType string

const (
	KeyTypeEC            KeyType = "EC"
	KeyTypeRSA           KeyType = "RSA"
	KeyTypeOctetSequence KeyType = "oct"
)

type Algorithm string

const (
	AlgorithmNone  Algorithm = ""
	AlgorithmHS256 Algorithm = "HS256"
	AlgorithmHS384 Algorithm = "HS384"
	AlgorithmHS512 Algorithm = "HS512"
)

var _ json.Marshaler = (*SymmetricKey)(nil)

type SymmetricKey struct {
	Algorithm Algorithm
	Key       []byte
}

// MarshalJSON implements json.Marshaler.
func (s SymmetricKey) MarshalJSON() ([]byte, error) {
	if s.Algorithm == AlgorithmNone {
		s.Algorithm = AlgorithmHS256
	}
	tmp := struct {
		KeyType   KeyType   `json:"kty"`
		Algorithm Algorithm `json:"alg"`
		Key       string    `json:"k"`
	}{
		KeyType:   KeyTypeOctetSequence,
		Algorithm: s.Algorithm,
		Key:       base64.RawURLEncoding.EncodeToString(s.Key),
	}

	return json.Marshal(tmp)
}
