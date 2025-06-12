package signature

type Signer interface {
	Sign(data []byte) ([]byte, error)
	Verify(data, signature []byte) bool
	PrivateKey() []byte
	PublicKey() []byte
}
