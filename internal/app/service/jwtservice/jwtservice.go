package jwtservice

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	logdoc "github.com/LogDoc-org/logdoc-go-appender/logrus"
	"github.com/golang-jwt/jwt"
	"github.com/gurkankaymak/hocon"
	"github.com/jmoiron/sqlx"
	"os"
	repository "sse-demo-core/internal/app/repository/users"
	"strings"
	"time"
)

type JwtServiceImpl struct {
	config *hocon.Config
	r      *repository.UserRepository
}

func New(config *hocon.Config, db *sqlx.DB) *JwtServiceImpl {
	urepo := repository.New(db)
	return &JwtServiceImpl{config: config, r: urepo}
}

func (s *JwtServiceImpl) ValidateToken(tokenStr string) (jwt.MapClaims, bool, error) {
	logger := logdoc.GetLogger()

	publicKey := readPublicPEMKey()

	// проверка токена
	tok, err := jwt.Parse(strings.ReplaceAll(tokenStr, "Bearer ", ""), func(jwtToken *jwt.Token) (interface{}, error) {
		if _, ok := jwtToken.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected method: %s", jwtToken.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		logger.Error("Ошибка формирования jwt токена, ", err.Error())
		return nil, false, err
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok || !tok.Valid {
		return nil, false, fmt.Errorf("invalid token, claims parse error: %w", err)
	}

	if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		return nil, false, fmt.Errorf("token expired")
	}

	if !claims.VerifyIssuer(s.config.GetString("jwt.issuer"), true) {
		return nil, false, fmt.Errorf("token issuer error")
	}

	if !claims.VerifyAudience(s.config.GetString("jwt.audience"), true) {
		return nil, false, fmt.Errorf("token audience error")
	}

	return claims, true, nil
}

func readPublicPEMKey() *rsa.PublicKey {
	logger := logdoc.GetLogger()

	// Читаем открытый ключ
	keyBytes, err := os.ReadFile("conf/keys/public.pem")
	if err != nil {
		logger.Error("Ошибка чтения открытого ключа, ", err.Error())
		return nil
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil {
		panic(err)
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		logger.Error("Ошибка парсинга открытого ключа, ", err.Error())
		return nil
	}

	switch t := publicKey.(type) {
	case *rsa.PublicKey:
		return t
	default:
		panic("unknown type of public key")
	}
}
