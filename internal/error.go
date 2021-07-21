package internal

type (
	// LockboxError are underling lockbox errors.
	LockboxError struct {
		message string
	}
)

// NewLockboxError will create a new lockbox error.
func NewLockboxError(message string) error {
	return &LockboxError{message: message}
}

func (err *LockboxError) Error() string {
	return err.message
}
