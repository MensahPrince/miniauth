package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// A simple not that passKey is just password

func BcryptHash(passKey string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(passKey), 14)
	return string(bytes), err
}

func BcryptCompareHash(passkey string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(passkey))
	return err == nil
}
