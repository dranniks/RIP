package session

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultRedisHost       = "localhost"
	defaultRedisPort       = "6379"
	defaultRedisPass       = "password"
	defaultRedisDB         = 0
	defaultSessionPrefix   = "lab4:sessions:"
	defaultSessionTTLMin   = 120
	defaultRedisTimeoutSec = 3
)

type Manager struct {
	addr        string
	password    string
	db          int
	keyPrefix   string
	sessionTTL  time.Duration
	dialTimeout time.Duration
}

type Record struct {
	UserID    uint   `json:"user_id"`
	Login     string `json:"login"`
	Role      string `json:"role"`
	CreatedAt int64  `json:"created_at"`
	ExpiresAt int64  `json:"expires_at"`
}

func NewManagerFromEnv() *Manager {
	host := envOrDefault("REDIS_HOST", defaultRedisHost)
	port := envOrDefault("REDIS_PORT", defaultRedisPort)
	password := strings.TrimSpace(os.Getenv("REDIS_PASSWORD"))
	if password == "" {
		password = defaultRedisPass
	}

	db := defaultRedisDB
	if rawDB := strings.TrimSpace(os.Getenv("REDIS_DB")); rawDB != "" {
		if parsed, err := strconv.Atoi(rawDB); err == nil && parsed >= 0 {
			db = parsed
		}
	}

	ttlMin := defaultSessionTTLMin
	if rawTTL := strings.TrimSpace(os.Getenv("SESSION_TTL_MINUTES")); rawTTL != "" {
		if parsed, err := strconv.Atoi(rawTTL); err == nil && parsed > 0 {
			ttlMin = parsed
		}
	}

	timeoutSec := defaultRedisTimeoutSec
	if rawTimeout := strings.TrimSpace(os.Getenv("REDIS_TIMEOUT_SECONDS")); rawTimeout != "" {
		if parsed, err := strconv.Atoi(rawTimeout); err == nil && parsed > 0 {
			timeoutSec = parsed
		}
	}

	prefix := strings.TrimSpace(os.Getenv("SESSION_KEY_PREFIX"))
	if prefix == "" {
		prefix = defaultSessionPrefix
	}

	return &Manager{
		addr:        net.JoinHostPort(strings.TrimSpace(host), strings.TrimSpace(port)),
		password:    password,
		db:          db,
		keyPrefix:   prefix,
		sessionTTL:  time.Duration(ttlMin) * time.Minute,
		dialTimeout: time.Duration(timeoutSec) * time.Second,
	}
}

func (m *Manager) SessionTTL() time.Duration {
	return m.sessionTTL
}

func (m *Manager) Key(sessionID string) string {
	return m.keyPrefix + strings.TrimSpace(sessionID)
}

func (m *Manager) Ping(ctx context.Context) error {
	resp, err := m.do(ctx, "PING")
	if err != nil {
		return err
	}

	value, ok := resp.(string)
	if !ok || strings.ToUpper(strings.TrimSpace(value)) != "PONG" {
		return fmt.Errorf("unexpected redis ping response: %v", resp)
	}
	return nil
}

func (m *Manager) CreateSession(ctx context.Context, userID uint, login string, role string) (string, time.Time, error) {
	if userID == 0 {
		return "", time.Time{}, fmt.Errorf("user id is required")
	}
	if strings.TrimSpace(login) == "" {
		return "", time.Time{}, fmt.Errorf("login is required")
	}
	if strings.TrimSpace(role) == "" {
		return "", time.Time{}, fmt.Errorf("role is required")
	}

	sessionID, err := generateSessionID()
	if err != nil {
		return "", time.Time{}, err
	}

	now := time.Now().UTC()
	expiresAt := now.Add(m.sessionTTL).UTC()
	record := Record{
		UserID:    userID,
		Login:     strings.TrimSpace(login),
		Role:      strings.TrimSpace(role),
		CreatedAt: now.Unix(),
		ExpiresAt: expiresAt.Unix(),
	}

	if err := m.SaveSession(ctx, sessionID, record, m.sessionTTL); err != nil {
		return "", time.Time{}, err
	}

	return sessionID, expiresAt, nil
}

func (m *Manager) SaveSession(ctx context.Context, sessionID string, record Record, ttl time.Duration) error {
	if strings.TrimSpace(sessionID) == "" {
		return fmt.Errorf("session id is required")
	}
	if ttl <= 0 {
		ttl = m.sessionTTL
	}
	if ttl <= 0 {
		return fmt.Errorf("invalid session ttl")
	}

	payload, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}

	seconds := int(ttl.Seconds())
	if seconds <= 0 {
		seconds = 1
	}

	resp, err := m.do(ctx, "SETEX", m.Key(sessionID), strconv.Itoa(seconds), string(payload))
	if err != nil {
		return err
	}

	value, ok := resp.(string)
	if !ok || strings.ToUpper(strings.TrimSpace(value)) != "OK" {
		return fmt.Errorf("unexpected redis SETEX response: %v", resp)
	}
	return nil
}

func (m *Manager) GetSession(ctx context.Context, sessionID string) (*Record, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, nil
	}

	resp, err := m.do(ctx, "GET", m.Key(sessionID))
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}

	rawPayload, ok := resp.(string)
	if !ok {
		return nil, fmt.Errorf("unexpected redis GET response: %T", resp)
	}

	record := Record{}
	if err := json.Unmarshal([]byte(rawPayload), &record); err != nil {
		return nil, fmt.Errorf("decode session payload: %w", err)
	}
	return &record, nil
}

func (m *Manager) DeleteSession(ctx context.Context, sessionID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return nil
	}

	_, err := m.do(ctx, "DEL", m.Key(sessionID))
	return err
}

func (m *Manager) do(ctx context.Context, args ...string) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("redis command is empty")
	}

	conn, err := net.DialTimeout("tcp", m.addr, m.dialTimeout)
	if err != nil {
		return nil, fmt.Errorf("redis connect %s: %w", m.addr, err)
	}
	defer conn.Close()

	deadline := time.Now().Add(m.dialTimeout)
	if ctx != nil {
		if dl, ok := ctx.Deadline(); ok && dl.Before(deadline) {
			deadline = dl
		}
	}
	_ = conn.SetDeadline(deadline)

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	if strings.TrimSpace(m.password) != "" {
		resp, err := m.exec(rw, "AUTH", m.password)
		if err != nil {
			return nil, fmt.Errorf("redis auth: %w", err)
		}
		if text, ok := resp.(string); !ok || strings.ToUpper(strings.TrimSpace(text)) != "OK" {
			return nil, fmt.Errorf("redis auth failed: %v", resp)
		}
	}

	if m.db > 0 {
		resp, err := m.exec(rw, "SELECT", strconv.Itoa(m.db))
		if err != nil {
			return nil, fmt.Errorf("redis select db: %w", err)
		}
		if text, ok := resp.(string); !ok || strings.ToUpper(strings.TrimSpace(text)) != "OK" {
			return nil, fmt.Errorf("redis select db failed: %v", resp)
		}
	}

	resp, err := m.exec(rw, args...)
	if err != nil {
		return nil, fmt.Errorf("redis %s: %w", strings.ToUpper(strings.TrimSpace(args[0])), err)
	}
	return resp, nil
}

func (m *Manager) exec(rw *bufio.ReadWriter, args ...string) (any, error) {
	if err := writeCommand(rw, args...); err != nil {
		return nil, err
	}
	if err := rw.Flush(); err != nil {
		return nil, err
	}
	return readResp(rw.Reader)
}

func writeCommand(rw *bufio.ReadWriter, args ...string) error {
	if _, err := rw.WriteString(fmt.Sprintf("*%d\r\n", len(args))); err != nil {
		return err
	}
	for _, arg := range args {
		if _, err := rw.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg)); err != nil {
			return err
		}
	}
	return nil
}

func readResp(r *bufio.Reader) (any, error) {
	prefix, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	switch prefix {
	case '+':
		line, err := readLine(r)
		if err != nil {
			return nil, err
		}
		return line, nil
	case '-':
		line, err := readLine(r)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf(strings.TrimSpace(line))
	case ':':
		line, err := readLine(r)
		if err != nil {
			return nil, err
		}
		value, err := strconv.ParseInt(strings.TrimSpace(line), 10, 64)
		if err != nil {
			return nil, err
		}
		return value, nil
	case '$':
		line, err := readLine(r)
		if err != nil {
			return nil, err
		}
		size, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil {
			return nil, err
		}
		if size == -1 {
			return nil, nil
		}
		payload := make([]byte, size+2)
		if _, err := io.ReadFull(r, payload); err != nil {
			return nil, err
		}
		return string(payload[:size]), nil
	case '*':
		line, err := readLine(r)
		if err != nil {
			return nil, err
		}
		count, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil {
			return nil, err
		}
		if count == -1 {
			return nil, nil
		}
		out := make([]any, 0, count)
		for i := 0; i < count; i++ {
			item, err := readResp(r)
			if err != nil {
				return nil, err
			}
			out = append(out, item)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported redis response prefix: %q", string(prefix))
	}
}

func readLine(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r"), nil
}

func generateSessionID() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate session id: %w", err)
	}
	return hex.EncodeToString(raw), nil
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
