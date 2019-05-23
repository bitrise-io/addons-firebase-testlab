package models

import (
	"encoding/json"
	"os"
	"time"

	"github.com/bitrise-io/go-crypto/crypto"
	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
	"github.com/pkg/errors"
)

// App ...
type App struct {
	ID                uuid.UUID `json:"id" db:"id"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
	Plan              string    `json:"plan" db:"plan"`
	EncryptedSecret   []byte    `json:"-" db:"encrypted_secret"`
	EncryptedSecretIV []byte    `json:"-" db:"encrypted_secret_iv"`
	AppSlug           string    `json:"app_slug" db:"app_slug"`
	BitriseAPIToken   string    `json:"-" db:"bitrise_api_token"` // to have authentication when making requests to Bitrise API
	APIToken          string    `json:"api_token" db:"api_token"` // to authenticate incoming requests from running builds
}

// String is not required by pop and may be deleted
func (a App) String() string {
	ja, err := json.Marshal(a)
	if err != nil {
		return ""
	}
	return string(ja)
}

// Apps is not required by pop and may be deleted
type Apps []App

// String is not required by pop and may be deleted
func (a Apps) String() string {
	ja, err := json.Marshal(a)
	if err != nil {
		return ""
	}
	return string(ja)
}

// Validate gets run everytime you call a "pop.Validate" method.
// This method is not required and may be deleted.
func (a *App) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: a.Plan, Name: "Plan"},
		&validators.StringIsPresent{Field: a.AppSlug, Name: "AppSlug"},
	), nil
}

// ValidateSave gets run everytime you call "pop.ValidateSave" method.
// This method is not required and may be deleted.
func (a *App) ValidateSave(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run everytime you call "pop.ValidateUpdate" method.
// This method is not required and may be deleted.
func (a *App) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// BeforeSave ...
func (a *App) BeforeSave(tx *pop.Connection) error {
	if len(a.EncryptedSecretIV) != 0 {
		return nil
	}

	var err error
	secret, err := crypto.SecureRandomHex(12)
	if err != nil {
		return errors.Wrap(err, "Failed to generate secret")
	}
	for {
		a.EncryptedSecretIV, err = crypto.GenerateIV()
		if err != nil {
			return errors.WithStack(err)
		}

		appWebhookCount, err := tx.Q().Where("apps.encrypted_secret_iv = ?", a.EncryptedSecretIV).Count(App{})
		if err != nil {
			return errors.WithStack(err)
		}
		if appWebhookCount == 0 {
			break
		}
	}

	err = a.encryptSecret(secret)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (a *App) encryptSecret(secret string) error {
	encryptKey, ok := os.LookupEnv("APP_WEBHOOK_SECRET_ENCRYPT_KEY")
	if !ok {
		return errors.New("No encrypt key provided")
	}
	encryptedSecret, err := crypto.AES256GCMCipher(secret, a.EncryptedSecretIV, encryptKey)
	if err != nil {
		return errors.WithStack(err)
	}
	a.EncryptedSecret = encryptedSecret

	return nil
}

// Secret ...
func (a *App) Secret() (string, error) {
	encryptKey, ok := os.LookupEnv("APP_WEBHOOK_SECRET_ENCRYPT_KEY")
	if !ok {
		return "", errors.New("No encrypt key provided")
	}
	secret, err := crypto.AES256GCMDecipher(a.EncryptedSecret, a.EncryptedSecretIV, encryptKey)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return secret, nil
}
