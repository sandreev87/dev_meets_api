package jwt

import (
	"dev_meets/internal/domain/models"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

const (
	SECRET = "hjqrhjqw124617ajfhajs"
)

func NewToken(user models.User, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(duration).Unix()

	tokenString, err := token.SignedString([]byte(SECRET))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyToken(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return "", fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(SECRET), nil
	})

	if err != nil {
		return 0, err
	}
	if !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	claims := token.Claims.(jwt.MapClaims)
	return int(claims["uid"].(float64)), nil
}
