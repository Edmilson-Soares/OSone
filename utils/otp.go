package utils

import (
	"fmt"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type AuthOPT struct {
	Secret string
	Url    string
}

func NewGenerate(email string) (*AuthOPT, error) {
	authOtp := AuthOPT{}
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "OSone",
		AccountName: email,
		SecretSize:  20,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
		Period:      30, // Janela de tempo de 30 segundos
	})
	if err != nil {
		return &authOtp, err
	}

	// Gerar um código TOTP
	now := time.Now()
	code, err := totp.GenerateCodeCustom(key.Secret(), now, totp.ValidateOpts{
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
		Period:    30,
	})
	if err != nil {
		return &authOtp, err
	}

	fmt.Printf("Código TOTP: %s\n", code)

	// Validar um código TOTP
	v, err := totp.ValidateCustom(code, key.Secret(), now, totp.ValidateOpts{
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
		Period:    30,
	})
	if err != nil {
		return &authOtp, err
	}
	fmt.Println(v)
	authOtp.Secret = key.Secret()
	authOtp.Url = key.URL()
	return &authOtp, nil

}

func NewValidateCode(secret, code string) (bool, error) {

	now := time.Now()
	return totp.ValidateCustom(code, secret, now, totp.ValidateOpts{
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
		Period:    30,
	})

}
