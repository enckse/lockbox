package encrypt

import (
	"crypto/sha512"
	"errors"
	random "math/rand"
	"time"

	"github.com/enckse/lockbox/internal/inputs"
	"golang.org/x/crypto/pbkdf2"
)

const (
	secretBoxAlgorithmVersion uint8 = 1
	isSecretBox                     = "secretbox"
	aesGCMAlgorithmVersion    uint8 = 2
)

type (
	algorithm interface {
		encrypt(k, d []byte) ([]byte, error)
		decrypt(k, d []byte) ([]byte, error)
		version() []byte
	}
)

func init() {
	random.Seed(time.Now().UnixNano())
}

func newAlgorithmFromVersion(vers uint8) algorithm {
	switch vers {
	case secretBoxAlgorithmVersion:
		return secretBoxAlgorithm{}
	case aesGCMAlgorithmVersion:
		return aesGCMAlgorithm{}
	}
	return nil
}

func newAlgorithm(mode string) algorithm {
	useMode := mode
	if mode == "" {
		useMode = inputs.EnvOrDefault(inputs.EncryptModeEnv, isSecretBox)
	}
	switch useMode {
	case isSecretBox:
		return secretBoxAlgorithm{}
	case "aes":
		return aesGCMAlgorithm{}
	}
	return nil
}

func algoVersion(v uint8) []byte {
	return []byte{0, v}
}

func pad(salt, key []byte) ([keyLength]byte, error) {
	d := pbkdf2.Key(key, salt, 4096, keyLength, sha512.New)
	if len(d) != keyLength {
		return [keyLength]byte{}, errors.New("invalid key result from pad")
	}
	var obj [keyLength]byte
	copy(obj[:], d[:keyLength])
	return obj, nil
}
