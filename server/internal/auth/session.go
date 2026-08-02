package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"finance-tracker/server/internal/store"
)

const SessionTTL = 30 * 24 * time.Hour

var ErrNoSession = errors.New("sessão inválida ou expirada")

type Sessions struct {
	Store *store.Store
}

func (s *Sessions) Create(ctx context.Context, userID int) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := hex.EncodeToString(raw)
	_, err := s.Store.Pool.Exec(ctx,
		`INSERT INTO sessions (token, user_id, expires_at) VALUES ($1, $2, $3)`,
		token, userID, time.Now().Add(SessionTTL))
	if err != nil {
		return "", err
	}
	return token, nil
}

// Lookup devolve o user_id da sessão e renova a validade (sessão deslizante).
func (s *Sessions) Lookup(ctx context.Context, token string) (int, error) {
	var userID int
	err := s.Store.Pool.QueryRow(ctx,
		`UPDATE sessions SET expires_at = $2 WHERE token = $1 AND expires_at > now() RETURNING user_id`,
		token, time.Now().Add(SessionTTL)).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNoSession
	}
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (s *Sessions) Delete(ctx context.Context, token string) error {
	_, err := s.Store.Pool.Exec(ctx, `DELETE FROM sessions WHERE token = $1`, token)
	return err
}
