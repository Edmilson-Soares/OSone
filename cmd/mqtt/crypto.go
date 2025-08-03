package mqtt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

// encrypt criptografa um texto usando AES-GCM
func Encrypt(texto string) (string, error) {
	plaintext := []byte(texto)

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

	// Criptografa o texto e prefixa com o nonce
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)

	// Converte para base64 para facilitar armazenamento/transmissão
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// decrypt descriptografa um texto previamente criptografado com AES-GCM
func Decrypt(textoCriptografado string) (string, error) {
	// Decodifica o texto em base64
	ciphertext, err := base64.URLEncoding.DecodeString(textoCriptografado)
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

	// Extrai o nonce do início do texto criptografado
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("texto criptografado muito curto")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Descriptografa o texto
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
