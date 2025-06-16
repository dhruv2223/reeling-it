package token

import (
	"os"

	"github.com/dhruv2223/reeling-it/logger"
)

type Token string

func GetJWTSecret(logger logger.Logger) Token {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-for-dev"
		logger.Info("JWT_SECRET not set, using default secret")
	} else {
		logger.Info("JWT_SECRET loaded from environment")
	}
	return Token(jwtSecret)
}
