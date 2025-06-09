package signature

import (
	"crypto/ed25519"
	"crypto/rand"
)

type EdDSASigner struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

func NewEdDSASigner() (*EdDSASigner, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &EdDSASigner{
		privateKey: priv,
		publicKey:  pub,
	}, nil
}

func (e *EdDSASigner) Sign(data []byte) ([]byte, error) {
	return ed25519.Sign(e.privateKey, data), nil
}

func (e *EdDSASigner) Verify(data, signature []byte) bool {
	return ed25519.Verify(e.publicKey, data, signature)
}

func (e *EdDSASigner) PrivateKey() []byte {
	return e.privateKey.Seed()
}

func (e *EdDSASigner) PublicKey() []byte {
	return e.publicKey
}
