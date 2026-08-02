package api

import (
	"context"
	"errors"
	"net/http"

	"finance-tracker/server/internal/auth"
	"finance-tracker/server/internal/store"
)

const sessionCookie = "ft_session"

type Server struct {
	Store         *store.Store
	Sessions      *auth.Sessions
	SecureCookies bool
}

func New(st *store.Store, secureCookies bool) *Server {
	return &Server{
		Store:         st,
		Sessions:      &auth.Sessions{Store: st},
		SecureCookies: secureCookies,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// públicos
	mux.HandleFunc("GET /api/setup", s.handleSetupStatus)
	mux.HandleFunc("POST /api/setup", s.handleSetup)
	mux.HandleFunc("POST /api/login", s.handleLogin)

	// autenticados
	mux.Handle("POST /api/logout", s.requireAuth(s.handleLogout))
	mux.Handle("GET /api/me", s.requireAuth(s.handleMe))

	mux.Handle("GET /api/categories", s.requireAuth(s.handleListCategories))
	mux.Handle("POST /api/categories", s.requireAuth(s.handleCreateCategory))
	mux.Handle("PUT /api/categories/{id}", s.requireAuth(s.handleUpdateCategory))
	mux.Handle("DELETE /api/categories/{id}", s.requireAuth(s.handleDeleteCategory))

	mux.Handle("GET /api/budgets", s.requireAuth(s.handleListBudgets))
	mux.Handle("PUT /api/budgets", s.requireAuth(s.handleSetBudget))

	mux.Handle("GET /api/expenses", s.requireAuth(s.handleListExpenses))
	mux.Handle("POST /api/expenses", s.requireAuth(s.handleCreateExpense))
	mux.Handle("PUT /api/expenses/{id}", s.requireAuth(s.handleUpdateExpense))
	mux.Handle("DELETE /api/expenses/{id}", s.requireAuth(s.handleDeleteExpense))

	mux.Handle("GET /api/months/{month}", s.requireAuth(s.handleMonthSummary))

	return mux
}

type ctxKey string

const userIDKey ctxKey = "userID"

func (s *Server) requireAuth(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookie)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "sessão em falta")
			return
		}
		userID, err := s.Sessions.Lookup(r.Context(), cookie.Value)
		if errors.Is(err, auth.ErrNoSession) {
			writeError(w, http.StatusUnauthorized, "sessão inválida ou expirada")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "erro interno")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userIDKey, userID)))
	})
}

func (s *Server) setSessionCookie(w http.ResponseWriter, token string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   s.SecureCookies,
		SameSite: http.SameSiteLaxMode,
	})
}
