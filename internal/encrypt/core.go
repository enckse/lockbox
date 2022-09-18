// Package encrypt handles encryption/decryption.
package encrypt

import (
	"crypto/rand"
	"crypto/sha512"
	"errors"
	"io"
	random "math/rand"
	"os"
	"time"

	"github.com/enckse/lockbox/internal/inputs"
	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/pbkdf2"
)

const (
	keyLength   = 32
	nonceLength = 24
	padLength   = 256
	saltLength  = 16
)

var (
	cryptoVers       = []byte{0, 1}
	cryptoVersLength = len(cryptoVers)
)

type (
	// Lockbox represents a method to encrypt/decrypt locked files.
	Lockbox struct {
		secret [keyLength]byte
		file   string
	}

	// LockboxOptions represent options to create a lockbox from.
	LockboxOptions struct {
		Key     string
		KeyMode string
		File    string
	}
)

// FromFile decrypts a file-system based encrypted file.
func FromFile(file string) ([]byte, error) {
	l, err := NewLockbox(LockboxOptions{File: file})
	if err != nil {
		return nil, err
	}
	return l.Decrypt()
}

// ToFile encrypts data to a file-system based file.
func ToFile(file string, data []byte) error {
	l, err := NewLockbox(LockboxOptions{File: file})
	if err != nil {
		return err
	}
	return l.Encrypt(data)
}

// NewLockbox creates a new usable lockbox instance.
func NewLockbox(options LockboxOptions) (Lockbox, error) {
	return newLockbox(options.Key, options.KeyMode, options.File)
}

func newLockbox(key, keyMode, file string) (Lockbox, error) {
	b, err := inputs.GetKey(key, keyMode)
	if err != nil {
		return Lockbox{}, err
	}
	var secretKey [keyLength]byte
	copy(secretKey[:], b)
	return Lockbox{secret: secretKey, file: file}, nil
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

func init() {
	random.Seed(time.Now().UnixNano())
}

// Encrypt will encrypt contents to file.
func (l Lockbox) Encrypt(datum []byte) error {
	var nonce [nonceLength]byte
	padTo := random.Intn(padLength)
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return err
	}
	data := datum
	if data == nil {
		b, err := inputs.RawStdin()
		if err != nil {
			return err
		}
		data = b
	}
	var padding [padLength]byte
	if _, err := io.ReadFull(rand.Reader, padding[:]); err != nil {
		return err
	}
	var salt [saltLength]byte
	if _, err := io.ReadFull(rand.Reader, salt[:]); err != nil {
		return err
	}
	var write []byte
	write = append(write, byte(padTo))
	write = append(write, padding[0:padTo]...)
	write = append(write, data...)
	key, err := pad(salt[:], l.secret[:])
	if err != nil {
		return err
	}
	encrypted := secretbox.Seal(nonce[:], write, &nonce, &key)
	var persist []byte
	persist = append(persist, cryptoVers...)
	persist = append(persist, salt[:]...)
	persist = append(persist, encrypted...)
	return os.WriteFile(l.file, persist, 0600)
}

// Decrypt will decrypt an object from file.
func (l Lockbox) Decrypt() ([]byte, error) {
	var nonce [nonceLength]byte
	var salt [saltLength]byte
	encrypted, err := os.ReadFile(l.file)
	if err != nil {
		return nil, err
	}
	copy(salt[:], encrypted[cryptoVersLength:saltLength+cryptoVersLength])
	copy(nonce[:], encrypted[cryptoVersLength+saltLength:cryptoVersLength+saltLength+nonceLength])
	key, err := pad(salt[:], l.secret[:])
	if err != nil {
		return nil, err
	}
	decrypted, ok := secretbox.Open(nil, encrypted[cryptoVersLength+saltLength+nonceLength:], &nonce, &key)
	if !ok {
		return nil, errors.New("decrypt not ok")
	}

	padding := int(decrypted[0])
	return decrypted[1+padding:], nil
}
