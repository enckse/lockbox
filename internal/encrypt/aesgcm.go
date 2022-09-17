package encrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

const (
	aesGCMAlgorithmSaltLength = 32
)

type (
	aesGCMAlgorithm struct {
	}
)

func (a aesGCMAlgorithm) version() algorithmVersions {
	return aesGCMAlgorithmVersion
}

func (a aesGCMAlgorithm) name() string {
	return "aesgcm"
}

func newCipher(key []byte, salt []byte) (cipher.Block, error) {
	useKey, err := pad(salt, key)
	if err != nil {
		return nil, err
	}
	return aes.NewCipher(useKey[:])
}

func (a aesGCMAlgorithm) encrypt(key, data []byte) ([]byte, error) {
	var salt [aesGCMAlgorithmSaltLength]byte
	if _, err := io.ReadFull(rand.Reader, salt[:]); err != nil {
		return nil, err
	}
	c, err := newCipher(key, salt[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	b := gcm.Seal(nonce, nonce, data, nil) //, nil
	var d []byte
	d = append(d, salt[:]...)
	d = append(d, b...)
	return d, nil
}

func (a aesGCMAlgorithm) decrypt(key, encrypted []byte) ([]byte, error) {
	var salt [aesGCMAlgorithmSaltLength]byte
	copy(salt[:], encrypted[0:aesGCMAlgorithmSaltLength])
	c, err := newCipher(key, salt[:])
	if err != nil {
		return nil, err
	}
	data := encrypted[aesGCMAlgorithmSaltLength:]
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	nonce := data[:nonceSize]
	datum := data[nonceSize:]

	return gcm.Open(nil, nonce, datum, nil)
}
