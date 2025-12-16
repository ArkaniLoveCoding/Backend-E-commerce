package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

func GenerateJwtToken(id uint, email string, role string) (string, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Gagal memuat data dari .env!")
	}
    secret_key := os.Getenv("JWTRAHASIA")
	if secret_key == "" {
		fmt.Println("Secret key jwt tidak ditemukan!")
	}
	var Secret_Key = []byte(secret_key)
	jwt := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": id,
		"email": email,
		"role": role,
		"exp": time.Now().Add(time.Hour * 180).Unix(),
	})
	tokenString, err := jwt.SignedString(Secret_Key)
	if err != nil {
		fmt.Println("Gagal melakukan decode jwt")
	}
	return tokenString, nil
}

func VerifyJwt(tokenString string) error {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Gagal memuat data dari .env!")
	}
	secret_key := os.Getenv("JWTRAHASIA")
	if secret_key == "" {
		return fmt.Errorf("Gagal menemukan code jwt!")
	}
	tokenVerify, err := jwt.Parse(tokenString, func (token *jwt.Token) (interface{}, error)  {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Gagal melakukan verify token!")
		}
		return []byte(secret_key), nil
	})
	if err != nil {
		fmt.Println(err.Error())
	}
	if !tokenVerify.Valid {
		return fmt.Errorf("Token yang dimasukkan tidak valid!")
	}
	return nil
}
func ExtractClaimsFromJWT(tokenString string) (jwt.MapClaims, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Gagal memuat data dari .env!")
	}
	jwtSecret := os.Getenv("JWTRAHASIA")
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("Metode signing tidak valid: %v", token.Header["alg"])
        }
        return []byte(jwtSecret), nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        return claims, nil
    }

    return nil, fmt.Errorf("Token tidak valid")
}
