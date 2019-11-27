package security

// JWTMock ...
type JWTMock struct {
	SignFn     func(authToken string) (string, error)
	VerifyFn   func(jwtToken string) (bool, error)
	GetTokenFn func(jwtToken string) (interface{}, error)
}

// Sign ...
func (j *JWTMock) Sign(authToken string) (string, error) {
	if j.SignFn == nil {
		panic("You have to override JWTService.Sign function in tests")
	}
	return j.SignFn(authToken)
}

// Verify ...
func (j *JWTMock) Verify(jwtToken string) (bool, error) {
	if j.VerifyFn == nil {
		panic("You have to override JWTService.Verify function in tests")
	}
	return j.VerifyFn(jwtToken)
}

// GetToken ...
func (j *JWTMock) GetToken(jwtToken string) (interface{}, error) {
	if j.GetTokenFn == nil {
		panic("You have to override JWTService.GetTokenFn function in tests")
	}
	return j.GetTokenFn(jwtToken)
}
