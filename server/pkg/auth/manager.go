package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// TokenManager provides logic for JWT & Refresh tokens generation and parsing.
type TokenManager interface {
	NewJWT(ttl time.Duration, userId string) (string, error)
	Parse(accessToken string) (CustomMapClaims, error)
	NewRefreshToken() (string, error)
}

type Manager struct {
	signingKey *rsa.PrivateKey
	verifyKey  *rsa.PublicKey
}

func NewManager(privateKeyPath *string, publicKeyPath string) (*Manager, error) {
	var privateKey *rsa.PrivateKey
	var err error
	// 私鑰只有Auth server有
	if privateKeyPath != nil && *privateKeyPath != "" {
		privateKey, err = parsePrivateKey(*privateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("error parsing private key: %v", err)
		}
	}

	publicKey, err := parsePublicKey(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("error parsing public key: %v", err)
	}

	return &Manager{signingKey: privateKey, verifyKey: publicKey}, nil
}

func parsePrivateKey(privateKeyPath string) (*rsa.PrivateKey, error) {
	pemData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, errors.New("failed to read PEM private file")
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func parsePublicKey(publicKeyPath string) (*rsa.PublicKey, error) {
	pemData, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read PEM public file: %w", err)
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PKIX public key: %w", err)
	}

	rsaPubKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not an RSA public key")
	}

	return rsaPubKey, nil
}

func (m *Manager) NewJWT(ttl time.Duration, userId string) (string, error) {
	// 通常StandardClaims是用sub，不過因為old peacock用的是user_id，所以改成MapClaims
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"exp":     time.Now().UTC().Add(ttl).Unix(),
		"user_id": userId,
	})

	return token.SignedString(m.signingKey)
}

func (m *Manager) Parse(accessToken string) (CustomMapClaims, error) {
	var respClaims CustomMapClaims
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (i interface{}, err error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return m.verifyKey, nil
	})
	if err != nil {
		if err.Error() == "crypto/rsa: verification error" {
			return respClaims, errors.New("invalid token")
		}
		return respClaims, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return respClaims, errors.New("error get user claims from token")
	}

	userId, ok := claims["user_id"].(string)
	if !ok {
		return respClaims, errors.New("error converting user_id to string")
	}

	respClaims.UserId = userId

	return respClaims, nil
}

func (m *Manager) NewRefreshToken() (string, error) {
	b := make([]byte, 32)

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	if _, err := r.Read(b); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", b), nil
}
