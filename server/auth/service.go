package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"

	"github.com/KasumiMercury/mock-todo-server/server/domain"
	"github.com/KasumiMercury/mock-todo-server/server/store"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type JWTKeyMode string

const (
	JWTKeyModeSecret JWTKeyMode = "secret"
	JWTKeyModeRSA    JWTKeyMode = "rsa"
)

type AuthService struct {
	userStore  store.UserStore
	keyMode    JWTKeyMode
	secretKey  []byte
	rsaPrivate *rsa.PrivateKey
	rsaPublic  *rsa.PublicKey
}

type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
	Kid string `json:"kid"`
}

type JWKSet struct {
	Keys []JWK `json:"keys"`
}

func NewAuthService(userStore store.UserStore, keyMode JWTKeyMode, secretKey string) (*AuthService, error) {
	service := &AuthService{
		userStore: userStore,
		keyMode:   keyMode,
	}

	switch keyMode {
	case JWTKeyModeSecret:
		service.secretKey = []byte(secretKey)
	case JWTKeyModeRSA:
		private, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA key: %w", err)
		}
		service.rsaPrivate = private
		service.rsaPublic = &private.PublicKey
	default:
		return nil, fmt.Errorf("unsupported key mode: %s", keyMode)
	}

	return service, nil
}

func (s *AuthService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func (s *AuthService) ValidatePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (s *AuthService) GenerateToken(user *domain.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"name": user.Username,
		"iat":  now.Unix(),
		"exp":  now.Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	switch s.keyMode {
	case JWTKeyModeSecret:
		return token.SignedString(s.secretKey)
	case JWTKeyModeRSA:
		token.Method = jwt.SigningMethodRS256
		return token.SignedString(s.rsaPrivate)
	default:
		return "", fmt.Errorf("unsupported key mode: %s", s.keyMode)
	}
}

func (s *AuthService) ValidateToken(tokenString string) (*jwt.Token, error) {
	switch s.keyMode {
	case JWTKeyModeSecret:
		return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return s.secretKey, nil
		})
	case JWTKeyModeRSA:
		return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return s.rsaPublic, nil
		})
	default:
		return nil, fmt.Errorf("unsupported key mode: %s", s.keyMode)
	}
}

func (s *AuthService) GetUserIDFromToken(token *jwt.Token) (int, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, fmt.Errorf("invalid token claims")
	}

	userID, ok := claims["sub"].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid user ID in token")
	}

	return int(userID), nil
}

func (s *AuthService) Login(username, password string) (*domain.User, string, error) {
	user, exists := s.userStore.GetByUsername(username)
	if !exists {
		return nil, "", fmt.Errorf("invalid credentials")
	}

	if !s.ValidatePassword(user.HashedPassword, password) {
		return nil, "", fmt.Errorf("invalid credentials")
	}

	token, err := s.GenerateToken(user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}

func (s *AuthService) Register(username, password string) (*domain.User, string, error) {
	// Check if user already exists
	if _, exists := s.userStore.GetByUsername(username); exists {
		return nil, "", fmt.Errorf("username already exists")
	}

	// Hash password
	hashedPassword, err := s.HashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &domain.User{
		Username:       username,
		HashedPassword: hashedPassword,
	}

	createdUser := s.userStore.Create(user)
	if createdUser == nil {
		return nil, "", fmt.Errorf("failed to create user")
	}

	// Generate token
	token, err := s.GenerateToken(createdUser)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return createdUser, token, nil
}

func (s *AuthService) GetJWKSet() (*JWKSet, error) {
	if s.keyMode != JWTKeyModeRSA {
		return nil, fmt.Errorf("JWKs only available in RSA mode")
	}

	// Convert RSA public key to JWK format
	n := base64.RawURLEncoding.EncodeToString(s.rsaPublic.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(s.rsaPublic.E)).Bytes())

	jwk := JWK{
		Kty: "RSA",
		Use: "sig",
		N:   n,
		E:   e,
		Kid: "rsa-key-1",
	}

	return &JWKSet{Keys: []JWK{jwk}}, nil
}
