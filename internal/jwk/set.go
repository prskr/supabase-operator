package jwk

type Key interface {
	SymmetricKey
}

type Set[T Key] struct {
	Keys []T `json:"keys"`
}
