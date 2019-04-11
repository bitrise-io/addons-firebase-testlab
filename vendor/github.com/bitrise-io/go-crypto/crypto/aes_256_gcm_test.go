package crypto_test

import (
	"testing"

	"github.com/bitrise-team/go-crypto/crypto"
	"github.com/stretchr/testify/require"
)

func Test_AES256GCMCipher(t *testing.T) {
	encryptKey := "619a5f2a096eac715901697daffaf571"
	textToEncrypt := "my-5up3r-s3cr3t"
	generatedIV, err := crypto.GenerateIV()
	require.NoError(t, err)

	t.Log("when all parameters have correct format and size")
	{
		encryptedText, err := crypto.AES256GCMCipher(textToEncrypt, generatedIV, encryptKey)
		require.NoError(t, err)
		require.Equal(t, len(textToEncrypt)+16, len(encryptedText))
	}

	t.Log("when encryt key doesn't have correct size")
	{
		encryptedText, err := crypto.AES256GCMCipher(textToEncrypt, generatedIV, "non correct key")
		require.Equal(t, "crypto/aes: invalid key size 15", err.Error())
		require.Equal(t, []byte(nil), encryptedText)
	}

	t.Log("when IV has wrong size")
	{
		require.Panics(t, func() {
			encryptedText, err := crypto.AES256GCMCipher(textToEncrypt, []byte("not correct IV"), encryptKey)
			require.NoError(t, err)
			require.Equal(t, "", encryptedText)
		})
	}
}

func Test_AES256GCMDecipher(t *testing.T) {
	encryptKey := "619a5f2a096eac715901697daffaf571"
	textToEncrypt := "my-5up3r-s3cr3t"
	generatedIV, err := crypto.GenerateIV()
	require.NoError(t, err)

	encryptedText, err := crypto.AES256GCMCipher(textToEncrypt, generatedIV, encryptKey)
	require.NoError(t, err)

	t.Log("when all parameters have correct format and size, retrieves the original value")
	{
		decryptedText, err := crypto.AES256GCMDecipher(encryptedText, generatedIV, encryptKey)
		require.NoError(t, err)
		require.Equal(t, textToEncrypt, decryptedText)
	}

	t.Log("when encryt key doesn't have correct size")
	{
		decryptedText, err := crypto.AES256GCMDecipher(encryptedText, generatedIV, "non correct key")
		require.Equal(t, "crypto/aes: invalid key size 15", err.Error())
		require.Equal(t, "", decryptedText)
	}
	otherValidEncryptkey := "157fb69958fe2aa3a38e46be6262d718"

	t.Log("when trying to decrypt with other valid encrypt key")
	{
		decryptedText, err := crypto.AES256GCMDecipher(encryptedText, generatedIV, otherValidEncryptkey)
		require.Equal(t, "cipher: message authentication failed", err.Error())
		require.Equal(t, "", decryptedText)
	}

	t.Log("when encrypted secret is belong to IV and encrypt key")
	{
		decryptedText, err := crypto.AES256GCMDecipher([]byte("non correct format text"), generatedIV, encryptKey)
		require.Equal(t, "cipher: message authentication failed", err.Error())
		require.Equal(t, "", decryptedText)
	}

	t.Log("when IV has wrong size")
	{
		require.NoError(t, err)
		require.Panics(t, func() {
			decryptedText, err := crypto.AES256GCMDecipher(encryptedText, []byte("not correct IV"), encryptKey)
			require.NoError(t, err)
			require.Equal(t, "", decryptedText)
		})
	}
}
