package crypto

import (
	"crypto/aes"
	"crypto/cipher"

	"github.com/pkg/errors"
)

// AES256GCMCipher ...
func AES256GCMCipher(textToEncrypt string, iv []byte, encryptKey string) ([]byte, error) {
	data := []byte(textToEncrypt)
	block, err := aes.NewCipher([]byte(encryptKey))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	ciphertext := aesgcm.Seal(nil, iv, data, nil)

	return ciphertext, nil
}

// AES256GCMDecipher ...
func AES256GCMDecipher(encryptedText, iv []byte, encryptKey string) (string, error) {
	block, err := aes.NewCipher([]byte(encryptKey))
	if err != nil {
		return "", errors.WithStack(err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.WithStack(err)
	}

	plaintext, err := aesgcm.Open(nil, iv, encryptedText, nil)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return string(plaintext), nil
}
