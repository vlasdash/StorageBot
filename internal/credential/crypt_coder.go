package credential

import (
	"github.com/AlexanderGrom/componenta/crypt"
)

type CryptCoder struct {
	secretKey string
}

func NewCryptCoder(key string) *CryptCoder {
	return &CryptCoder{
		secretKey: key,
	}
}

func (c *CryptCoder) Encrypt(password string) (string, error) {
	return crypt.Encrypt(password, c.secretKey)
}

func (c *CryptCoder) Decrypt(passwordCrypt string) (string, error) {
	return crypt.Decrypt(passwordCrypt, c.secretKey)
}
