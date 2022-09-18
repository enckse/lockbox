package encrypt

import (
	"crypto/rand"
	"errors"
	"io"

	"golang.org/x/crypto/nacl/secretbox"
)

type (
	secretBoxAlgorithm struct {
	}
)

const (
	secretBoxAlgorithmNonceLength = 24
	secretBoxAlgorithmSaltLength  = 16
)

func (s secretBoxAlgorithm) dataSize() int {
	return secretBoxAlgorithmSaltLength + secretBoxAlgorithmNonceLength
}

func (s secretBoxAlgorithm) name() string {
	return "secretbox"
}

func (s secretBoxAlgorithm) version() algorithmVersions {
	return secretBoxAlgorithmVersion
}

func (s secretBoxAlgorithm) encrypt(encryptKey, data []byte) ([]byte, error) {
	var nonce [secretBoxAlgorithmNonceLength]byte
	var salt [secretBoxAlgorithmSaltLength]byte
	if _, err := io.ReadFull(rand.Reader, salt[:]); err != nil {
		return nil, err
	}
	key, err := pad(salt[:], encryptKey[:])
	if err != nil {
		return nil, err
	}
	encrypted := secretbox.Seal(nonce[:], data, &nonce, &key)
	var persist []byte
	persist = append(persist, salt[:]...)
	persist = append(persist, encrypted...)
	return persist, nil
}

func (s secretBoxAlgorithm) decrypt(encryptKey, encrypted []byte) ([]byte, error) {
	var nonce [secretBoxAlgorithmNonceLength]byte
	var salt [secretBoxAlgorithmSaltLength]byte
	copy(salt[:], encrypted[0:secretBoxAlgorithmSaltLength])
	copy(nonce[:], encrypted[secretBoxAlgorithmSaltLength:secretBoxAlgorithmSaltLength+secretBoxAlgorithmNonceLength])
	key, err := pad(salt[:], encryptKey[:])
	if err != nil {
		return nil, err
	}
	decrypted, ok := secretbox.Open(nil, encrypted[secretBoxAlgorithmSaltLength+secretBoxAlgorithmNonceLength:], &nonce, &key)
	if !ok {
		return nil, errors.New("decrypt not ok")
	}

	return decrypted, nil
}
