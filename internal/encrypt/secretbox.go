package encrypt

import (
	"crypto/rand"
	"errors"
	"io"
	random "math/rand"

	"golang.org/x/crypto/nacl/secretbox"
)

type (
	secretBoxAlgorithm struct {
	}
)

const (
	secretBoxAlgorithmNonceLength = 24
	secretBoxAlgorithmPadLength   = 256
	secretBoxAlgorithmSaltLength  = 16
)

func (s secretBoxAlgorithm) version() []byte {
	return algoVersion(secretBoxAlgorithmVersion)
}

func (s secretBoxAlgorithm) encrypt(encryptKey, data []byte) ([]byte, error) {
	var nonce [secretBoxAlgorithmNonceLength]byte
	padTo := random.Intn(secretBoxAlgorithmPadLength)
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nil, err
	}
	var padding [secretBoxAlgorithmPadLength]byte
	if _, err := io.ReadFull(rand.Reader, padding[:]); err != nil {
		return nil, err
	}
	var salt [secretBoxAlgorithmSaltLength]byte
	if _, err := io.ReadFull(rand.Reader, salt[:]); err != nil {
		return nil, err
	}
	var write []byte
	write = append(write, byte(padTo))
	write = append(write, padding[0:padTo]...)
	write = append(write, data...)
	key, err := pad(salt[:], encryptKey[:])
	if err != nil {
		return nil, err
	}
	encrypted := secretbox.Seal(nonce[:], write, &nonce, &key)
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

	padding := int(decrypted[0])
	return decrypted[1+padding:], nil
}
