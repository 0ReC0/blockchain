package encryption

type Encryptor interface {
	Encrypt(plaintext, key []byte) ([]byte, error)
	Decrypt(ciphertext, key []byte) ([]byte, error)
}
