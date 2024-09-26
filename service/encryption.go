package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

type EncryptionService interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}

type encryptionService struct {
	password string
}

func NewEncryptionService(password string) EncryptionService {
	// Pad or truncate the password to 32 bytes
	if len(password) < 32 {
		password = password + string(make([]byte, 32-len(password)))
	} else {
		password = password[:32]
	}
	return &encryptionService{password}
}

func (e *encryptionService) Encrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher([]byte(e.password))

	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(data))
	iv := ciphertext[:aes.BlockSize]

	_, err = io.ReadFull(rand.Reader, iv)

	if err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], data)

	return ciphertext, nil
}

func (e *encryptionService) Decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher([]byte(e.password))

	if err != nil {
		return nil, err
	}

	if len(data) < aes.BlockSize {
		return nil, err
	}

	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(data, data)

	return data, nil
}
