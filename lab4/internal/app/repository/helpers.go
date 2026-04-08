package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"
)

func ctxTimeout() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
	return ctx
}

func hashPassword(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func stringPtr(v string) *string {
	value := v
	return &value
}
