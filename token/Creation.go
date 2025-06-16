package token

import (
	"time"

	"github.com/dhruv2223/reeling-it/logger"
	"github.com/dhruv2223/reeling-it/models"
	"github.com/golang-jwt/jwt/v5"
)

func CreateJWT(user models.User, logger logger.Logger) Token {
	jwtSecret := GetJWTSecret(logger)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
		"exp":   time.Now().Add(time.Hour * 72).Unix(),
	})
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		logger.Error("Failed to sign JWT", err)
		return Token("")
	}
	return Token(tokenString)
}
