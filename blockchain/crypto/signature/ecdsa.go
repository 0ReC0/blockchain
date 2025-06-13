package signature

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"fmt"
	"math/big"
)

type ECDSASigner struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
}

func NewSignerFromKey(privateKey *ecdsa.PrivateKey) *ECDSASigner {
	return &ECDSASigner{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}
}

func NewECDSASigner() (*ECDSASigner, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return &ECDSASigner{
		privateKey: privKey,
		publicKey:  &privKey.PublicKey,
	}, nil
}

type ecdsaSignature struct {
	R, S *big.Int
}

func (e *ECDSASigner) Sign(data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, e.privateKey, hash[:])
	if err != nil {
		return nil, err
	}

	// Сериализуем подпись в DER-формат
	return asn1.Marshal(ecdsaSignature{R: r, S: s})
}

func (e *ECDSASigner) Verify(data, signature []byte) bool {
	hash := sha256.Sum256(data)

	// Десериализуем подпись
	sig := new(ecdsaSignature)
	_, err := asn1.Unmarshal(signature, sig)
	if err != nil {
		return false
	}

	return ecdsa.Verify(e.publicKey, hash[:], sig.R, sig.S)
}

func (e *ECDSASigner) PrivateKey() []byte {
	return e.privateKey.D.Bytes()
}

func (e *ECDSASigner) PublicKey() []byte {
	return elliptic.Marshal(elliptic.P256(), e.publicKey.X, e.publicKey.Y)
}

func ParsePublicKey(data []byte) (*ecdsa.PublicKey, error) {
	if len(data) != 65 {
		return nil, fmt.Errorf("invalid public key length: expected 65 bytes, got %d", len(data))
	}

	if data[0] != 0x04 {
		return nil, fmt.Errorf("invalid public key format: expected uncompressed format (0x04)")
	}

	x := new(big.Int).SetBytes(data[1:33])
	y := new(big.Int).SetBytes(data[33:65])

	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}, nil
}
