package security

import (
	"crypto/rsa"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

// JWTInterface ...
type JWTInterface interface {
	Sign(authToken string) (string, error)
	Verify(jwtToken string) (bool, error)
	GetToken(jwtToken string) (interface{}, error)
}

// JWTService ...
type JWTService struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	expiration time.Duration
}

// NewJWTService ...
func NewJWTService(publicKey, privateKey string, expiration time.Duration) (JWTService, error) {
	signKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
	if err != nil {
		return JWTService{}, errors.Wrap(err, "Failed to parse private key")
	}
	verifyKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKey))
	if err != nil {
		return JWTService{}, errors.Wrap(err, "Failed to parse public key")
	}
	return JWTService{
		privateKey: signKey,
		publicKey:  verifyKey,
		expiration: expiration,
	}, nil
}

// Sign ...
func (j *JWTService) Sign(authToken string) (string, error) {
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"token": authToken,
		"exp":   time.Now().Add(j.expiration).Unix(),
	})

	return jwtToken.SignedString(j.privateKey)
}

// Verify ...
func (j *JWTService) Verify(jwtToken string) (bool, error) {
	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		return j.publicKey, nil
	})
	if err != nil {
		return false, err
	}
	return token.Valid, nil
}

// GetToken ...
func (j *JWTService) GetToken(jwtToken string) (interface{}, error) {
	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		return j.publicKey, nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims["token"], nil
	}
	return "", errors.New("Token is not valid")
}
