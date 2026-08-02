package api

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"finance-tracker/server/internal/auth"
)

// handleSetupStatus indica se a app ainda precisa do registo inicial.
func (s *Server) handleSetupStatus(w http.ResponseWriter, r *http.Request) {
	var n int
	if err := s.Store.Pool.QueryRow(r.Context(), `SELECT count(*) FROM users`).Scan(&n); err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"needsSetup": n == 0})
}

type setupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// handleSetup cria o único utilizador da app; só funciona enquanto não existir nenhum.
func (s *Server) handleSetup(w http.ResponseWriter, r *http.Request) {
	var req setupRequest
	if !readJSON(w, r, &req) {
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" {
		req.Username = "loureiro"
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "a password deve ter pelo menos 8 caracteres")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	var userID int
	err = s.Store.Pool.QueryRow(r.Context(),
		`INSERT INTO users (username, password_hash)
		 SELECT $1, $2 WHERE NOT EXISTS (SELECT 1 FROM users)
		 RETURNING id`, req.Username, hash).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusConflict, "a app já foi configurada")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	token, err := s.Sessions.Create(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	s.setSessionCookie(w, token, int(auth.SessionTTL/time.Second))
	writeJSON(w, http.StatusCreated, map[string]any{"username": req.Username})
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !readJSON(w, r, &req) {
		return
	}

	var userID int
	var username, hash string
	err := s.Store.Pool.QueryRow(r.Context(),
		`SELECT id, username, password_hash FROM users WHERE username = $1`,
		strings.TrimSpace(req.Username)).Scan(&userID, &username, &hash)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && !auth.VerifyPassword(req.Password, hash)) {
		writeError(w, http.StatusUnauthorized, "credenciais inválidas")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	token, err := s.Sessions.Create(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	s.setSessionCookie(w, token, int(auth.SessionTTL/time.Second))
	writeJSON(w, http.StatusOK, map[string]any{"username": username})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookie); err == nil {
		s.Sessions.Delete(r.Context(), cookie.Value)
	}
	s.setSessionCookie(w, "", -1)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userIDKey).(int)
	var username string
	if err := s.Store.Pool.QueryRow(r.Context(),
		`SELECT username FROM users WHERE id = $1`, userID).Scan(&username); err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"username": username})
}
