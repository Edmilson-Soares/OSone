package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"strings"
)

// encrypt criptografa um text usando AES-GCM
func Encrypt(text string) (string, error) {
	plaintext := []byte(text)

	block, err := aes.NewCipher([]byte(os.Getenv("SK_CRYPTO")))
	if err != nil {
		return "", err
	}

	// GCM (Galois/Counter Mode) fornece autenticação e criptografia
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Cria um nonce (número usado uma vez) de tamanho adequado
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Criptografa o text e prefixa com o nonce
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)

	// Converte para base64 para facilitar armazenamento/transmissão
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// decrypt descriptografa um text previamente criptografado com AES-GCM
func Decrypt(textCriptografado string) (string, error) {
	// Decodifica o text em base64
	text := strings.ReplaceAll(textCriptografado, "%3D", "=")
	ciphertext, err := base64.URLEncoding.DecodeString(text)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(os.Getenv("SK_CRYPTO")))
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Extrai o nonce do início do text criptografado
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("text criptografado muito curto")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Descriptografa o text
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func GenerateKey() ([]byte, error) {
	key := make([]byte, 32) // 32 bytes para AES-256
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}
