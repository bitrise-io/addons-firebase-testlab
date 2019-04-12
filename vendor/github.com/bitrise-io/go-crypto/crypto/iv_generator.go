package crypto

import "github.com/pkg/errors"

// GenerateIV ...
func GenerateIV() ([]byte, error) {
	secureRandomBytes, err := SecureRandomBytes(12)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return secureRandomBytes, nil
}
