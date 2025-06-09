package hash

type Hasher interface {
	Hash(data []byte) []byte
	Name() string
}
