package helper

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userID int64, role string, companyID int64) (string, error) {
	claims := jwt.MapClaims{
		"user_id":    userID,
		"role":       role,
		"company_id": companyID,
		"exp":        time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(ENV("JWT_KEY")))
}
