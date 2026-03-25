package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultJWTSecret     = "lab4-jwt-secret-change-me"
	defaultJWTTTLMinutes = 120
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token is expired")
)

type Claims struct {
	UserID    uint   `json:"uid"`
	Login     string `json:"login"`
	Role      string `json:"role"`
	SessionID string `json:"sid"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

type Manager struct {
	secret []byte
	ttl    time.Duration
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

func NewManagerFromEnv() *Manager {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		secret = defaultJWTSecret
	}

	ttlMinutes := defaultJWTTTLMinutes
	rawTTL := strings.TrimSpace(os.Getenv("JWT_TTL_MINUTES"))
	if rawTTL != "" {
		if parsed, err := strconv.Atoi(rawTTL); err == nil && parsed > 0 {
			ttlMinutes = parsed
		}
	}

	return &Manager{
		secret: []byte(secret),
		ttl:    time.Duration(ttlMinutes) * time.Minute,
	}
}

func (m *Manager) TTL() time.Duration {
	return m.ttl
}

func (m *Manager) IssueToken(userID uint, login string, role string, sessionID string) (string, *Claims, error) {
	if userID == 0 {
		return "", nil, fmt.Errorf("%w: user id is required", ErrInvalidToken)
	}
	if strings.TrimSpace(login) == "" {
		return "", nil, fmt.Errorf("%w: login is required", ErrInvalidToken)
	}
	if strings.TrimSpace(role) == "" {
		return "", nil, fmt.Errorf("%w: role is required", ErrInvalidToken)
	}
	if strings.TrimSpace(sessionID) == "" {
		return "", nil, fmt.Errorf("%w: session id is required", ErrInvalidToken)
	}

	now := time.Now().UTC()
	claims := &Claims{
		UserID:    userID,
		Login:     strings.TrimSpace(login),
		Role:      strings.TrimSpace(role),
		SessionID: strings.TrimSpace(sessionID),
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(m.ttl).Unix(),
	}

	token, err := m.buildSignedToken(claims)
	if err != nil {
		return "", nil, err
	}

	return token, claims, nil
}

func (m *Manager) ParseToken(rawToken string) (*Claims, error) {
	token := strings.TrimSpace(rawToken)
	if token == "" {
		return nil, fmt.Errorf("%w: token is empty", ErrInvalidToken)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("%w: token format is invalid", ErrInvalidToken)
	}

	signingInput := parts[0] + "." + parts[1]
	expectedSig, err := m.sign(signingInput)
	if err != nil {
		return nil, err
	}
	if subtle.ConstantTimeCompare([]byte(expectedSig), []byte(parts[2])) != 1 {
		return nil, fmt.Errorf("%w: bad signature", ErrInvalidToken)
	}

	headerData, err := decodeBase64URL(parts[0])
	if err != nil {
		return nil, fmt.Errorf("%w: bad header", ErrInvalidToken)
	}
	header := jwtHeader{}
	if err := json.Unmarshal(headerData, &header); err != nil {
		return nil, fmt.Errorf("%w: bad header json", ErrInvalidToken)
	}
	if header.Alg != "HS256" || header.Typ != "JWT" {
		return nil, fmt.Errorf("%w: unsupported token header", ErrInvalidToken)
	}

	payloadData, err := decodeBase64URL(parts[1])
	if err != nil {
		return nil, fmt.Errorf("%w: bad payload", ErrInvalidToken)
	}
	claims := Claims{}
	if err := json.Unmarshal(payloadData, &claims); err != nil {
		return nil, fmt.Errorf("%w: bad payload json", ErrInvalidToken)
	}

	nowUnix := time.Now().UTC().Unix()
	if claims.ExpiresAt <= nowUnix {
		return nil, ErrExpiredToken
	}
	if claims.UserID == 0 || strings.TrimSpace(claims.Login) == "" || strings.TrimSpace(claims.Role) == "" || strings.TrimSpace(claims.SessionID) == "" {
		return nil, fmt.Errorf("%w: payload fields are incomplete", ErrInvalidToken)
	}

	return &claims, nil
}

func ExtractBearerToken(header string) (string, error) {
	value := strings.TrimSpace(header)
	if value == "" {
		return "", fmt.Errorf("%w: authorization header is empty", ErrInvalidToken)
	}

	parts := strings.SplitN(value, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(strings.TrimSpace(parts[0]), "Bearer") {
		return "", fmt.Errorf("%w: expected Authorization: Bearer <token>", ErrInvalidToken)
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", fmt.Errorf("%w: bearer token is empty", ErrInvalidToken)
	}

	return token, nil
}

func (m *Manager) buildSignedToken(claims *Claims) (string, error) {
	headerBytes, err := json.Marshal(jwtHeader{Alg: "HS256", Typ: "JWT"})
	if err != nil {
		return "", fmt.Errorf("marshal jwt header: %w", err)
	}
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal jwt payload: %w", err)
	}

	headerPart := encodeBase64URL(headerBytes)
	payloadPart := encodeBase64URL(payloadBytes)
	signingInput := headerPart + "." + payloadPart

	signature, err := m.sign(signingInput)
	if err != nil {
		return "", err
	}

	return signingInput + "." + signature, nil
}

func (m *Manager) sign(value string) (string, error) {
	if len(m.secret) == 0 {
		return "", fmt.Errorf("%w: jwt secret is empty", ErrInvalidToken)
	}

	mac := hmac.New(sha256.New, m.secret)
	if _, err := mac.Write([]byte(value)); err != nil {
		return "", fmt.Errorf("jwt sign: %w", err)
	}
	return encodeBase64URL(mac.Sum(nil)), nil
}

func encodeBase64URL(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func decodeBase64URL(value string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(value)
}
