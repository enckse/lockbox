package internal

import (
	"crypto/rand"
	"io"
	random "math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/shlex"
	"golang.org/x/crypto/nacl/secretbox"
)

const (
	keyLength   = 32
	nonceLength = 24
	padLength   = 256
	// PlainKeyMode is plaintext based key resolution.
	PlainKeyMode = "plaintext"
	// CommandKeyMode will run an external command to get the key (from stdout).
	CommandKeyMode = "command"
)

type (
	// Lockbox represents a method to encrypt/decrypt locked files.
	Lockbox struct {
		secret [keyLength]byte
		file   string
	}
)

// NewLockbox creates a new lockbox for encryption/decryption.
func NewLockbox(key, keyMode, file string) (Lockbox, error) {
	useKeyMode := keyMode
	if useKeyMode == "" {
		useKeyMode = os.Getenv("LOCKBOX_KEYMODE")
	}
	if useKeyMode == "" {
		useKeyMode = CommandKeyMode
	}
	useKey := key
	if useKey == "" {
		useKey = os.Getenv("LOCKBOX_KEY")
	}
	b, err := getKey(useKeyMode, useKey)
	if err != nil {
		return Lockbox{}, err
	}

	if len(b) == 0 {
		return Lockbox{}, NewLockboxError("key is empty")
	}

	if len(b) > keyLength {
		return Lockbox{}, NewLockboxError("key is too large for use")
	}

	for len(b) < keyLength {
		b = append(b, byte(0))
	}
	var secretKey [keyLength]byte
	copy(secretKey[:], b)
	return Lockbox{secret: secretKey, file: file}, nil
}

func getKey(keyMode, name string) ([]byte, error) {
	var data []byte
	switch keyMode {
	case CommandKeyMode:
		parts, err := shlex.Split(name)
		if err != nil {
			return nil, err
		}
		cmd := exec.Command(parts[0], parts[1:]...)
		b, err := cmd.Output()
		if err != nil {
			return nil, err
		}
		data = b
	case PlainKeyMode:
		data = []byte(name)
	default:
		return nil, NewLockboxError("unknown keymode")
	}
	return []byte(strings.TrimSpace(string(data))), nil
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
		b, err := getStdin(false)
		if err != nil {
			return err
		}
		data = b
	}
	var padding [padLength]byte
	if _, err := io.ReadFull(rand.Reader, padding[:]); err != nil {
		return err
	}
	var write []byte
	write = append(write, byte(padTo))
	write = append(write, padding[0:padTo]...)
	write = append(write, data...)
	encrypted := secretbox.Seal(nonce[:], write, &nonce, &l.secret)
	return os.WriteFile(l.file, encrypted, 0600)
}

// Decrypt will decrypt an object from file.
func (l Lockbox) Decrypt() ([]byte, error) {
	var nonce [nonceLength]byte
	encrypted, err := os.ReadFile(l.file)
	if err != nil {
		return nil, err
	}
	copy(nonce[:], encrypted[:nonceLength])
	decrypted, ok := secretbox.Open(nil, encrypted[nonceLength:], &nonce, &l.secret)
	if !ok {
		return nil, NewLockboxError("decrypt not ok")
	}

	padding := int(decrypted[0])
	return decrypted[1+padding:], nil
}
