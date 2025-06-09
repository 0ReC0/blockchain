package signature

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"math/big"
)

type ECDSASigner struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
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

func (e *ECDSASigner) Sign(data []byte) ([]byte, error) {
	rand := rand.Reader
	hash := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand, e.privateKey, hash[:])
	if err != nil {
		return nil, err
	}
	params := e.publicKey.Curve.Params()
	curveBits := params.BitSize
	rBytes, sBytes := r.Bytes(), s.Bytes()
	signature := make([]byte, 2*curveBits/8)
	copy(signature[1*curveBits/8-len(rBytes):], rBytes)
	copy(signature[2*curveBits/8-len(sBytes):], sBytes)
	return signature, nil
}

func (e *ECDSASigner) Verify(data, signature []byte) bool {
	hash := sha256.Sum256(data)
	curveBits := e.publicKey.Curve.Params().BitSize
	rBytes := signature[:curveBits/8]
	sBytes := signature[curveBits/8:]
	r := new(big.Int).SetBytes(rBytes)
	s := new(big.Int).SetBytes(sBytes)
	return ecdsa.Verify(e.publicKey, hash[:], r, s)
}

func (e *ECDSASigner) PrivateKey() []byte {
	return e.privateKey.D.Bytes()
}

func (e *ECDSASigner) PublicKey() []byte {
	return elliptic.Marshal(elliptic.P256(), e.publicKey.X, e.publicKey.Y)
}
