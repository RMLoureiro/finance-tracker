package api

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Category struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Icon     string `json:"icon"`
	Archived bool   `json:"archived"`
	Position int    `json:"position"`
}

var colorRe = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

func (s *Server) handleListCategories(w http.ResponseWriter, r *http.Request) {
	rows, err := s.Store.Pool.Query(r.Context(),
		`SELECT id, name, color, icon, archived, position FROM categories ORDER BY position, id`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	defer rows.Close()
	cats := []Category{}
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Color, &c.Icon, &c.Archived, &c.Position); err != nil {
			writeError(w, http.StatusInternalServerError, "erro interno")
			return
		}
		cats = append(cats, c)
	}
	writeJSON(w, http.StatusOK, cats)
}

type categoryRequest struct {
	Name     string `json:"name"`
	Color    string `json:"color"`
	Icon     string `json:"icon"`
	Archived bool   `json:"archived"`
	Position *int   `json:"position"`
}

func (req *categoryRequest) validate(w http.ResponseWriter) bool {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "o nome da categoria é obrigatório")
		return false
	}
	if !colorRe.MatchString(req.Color) {
		writeError(w, http.StatusBadRequest, "cor inválida (esperado #RRGGBB)")
		return false
	}
	return true
}

func (s *Server) handleCreateCategory(w http.ResponseWriter, r *http.Request) {
	var req categoryRequest
	if !readJSON(w, r, &req) || !req.validate(w) {
		return
	}
	var c Category
	err := s.Store.Pool.QueryRow(r.Context(),
		`INSERT INTO categories (name, color, icon, position)
		 VALUES ($1, $2, $3, COALESCE((SELECT max(position) + 1 FROM categories), 0))
		 RETURNING id, name, color, icon, archived, position`,
		req.Name, req.Color, req.Icon).Scan(&c.ID, &c.Name, &c.Color, &c.Icon, &c.Archived, &c.Position)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func (s *Server) handleUpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	var req categoryRequest
	if !readJSON(w, r, &req) || !req.validate(w) {
		return
	}
	var c Category
	err = s.Store.Pool.QueryRow(r.Context(),
		`UPDATE categories
		 SET name = $2, color = $3, icon = $4, archived = $5, position = COALESCE($6, position)
		 WHERE id = $1
		 RETURNING id, name, color, icon, archived, position`,
		id, req.Name, req.Color, req.Icon, req.Archived, req.Position).
		Scan(&c.ID, &c.Name, &c.Color, &c.Icon, &c.Archived, &c.Position)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "categoria não encontrada")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (s *Server) handleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	tag, err := s.Store.Pool.Exec(r.Context(), `DELETE FROM categories WHERE id = $1`, id)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		writeError(w, http.StatusConflict, "a categoria tem despesas associadas — arquiva-a em vez de a apagar")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "categoria não encontrada")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
