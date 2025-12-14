package utils

import (
	"golang.org/x/crypto/bcrypt"
)

func GenerateAndHashPassword (password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CompareAndHashPassword (hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return  err 
}