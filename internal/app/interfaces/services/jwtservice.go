package services

import "github.com/golang-jwt/jwt"

type JwtService interface {
	ValidateToken(tokenStr string) (jwt.MapClaims, bool, error)
}
