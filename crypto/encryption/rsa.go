package encryption

import (
	"crypto/rand"
	"crypto/rsa"
)

type RSAEncryptor struct {
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}

func NewRSAEncryptor() (*RSAEncryptor, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return &RSAEncryptor{
		PublicKey:  &privKey.PublicKey,
		PrivateKey: privKey,
	}, nil
}

func (r *RSAEncryptor) Encrypt(plaintext []byte, _ []byte) ([]byte, error) {
	return rsa.EncryptPKCS1v15(rand.Reader, r.PublicKey, plaintext)
}

func (r *RSAEncryptor) Decrypt(ciphertext []byte, _ []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, r.PrivateKey, ciphertext)
}
