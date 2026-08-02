package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type Recurring struct {
	ID                  int    `json:"id"`
	CategoryID          int    `json:"categoryId"`
	Description         string `json:"description"`
	ExpectedAmountCents int64  `json:"expectedAmountCents"`
	Frequency           string `json:"frequency"`
	DueDay              *int   `json:"dueDay"`
	DueMonth            *int   `json:"dueMonth"`
	Active              bool   `json:"active"`
}

const recurringCols = `id, category_id, description, expected_amount_cents, frequency, due_day, due_month, active`

func scanRecurring(row pgx.Row) (Recurring, error) {
	var rec Recurring
	err := row.Scan(&rec.ID, &rec.CategoryID, &rec.Description, &rec.ExpectedAmountCents,
		&rec.Frequency, &rec.DueDay, &rec.DueMonth, &rec.Active)
	return rec, err
}

func (s *Server) handleListRecurring(w http.ResponseWriter, r *http.Request) {
	rows, err := s.Store.Pool.Query(r.Context(),
		`SELECT `+recurringCols+` FROM recurring_expenses ORDER BY id`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	defer rows.Close()
	list := []Recurring{}
	for rows.Next() {
		rec, err := scanRecurring(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "erro interno")
			return
		}
		list = append(list, rec)
	}
	writeJSON(w, http.StatusOK, list)
}

type recurringRequest struct {
	CategoryID          int    `json:"categoryId"`
	Description         string `json:"description"`
	ExpectedAmountCents int64  `json:"expectedAmountCents"`
	Frequency           string `json:"frequency"`
	DueDay              *int   `json:"dueDay"`
	DueMonth            *int   `json:"dueMonth"`
	Active              *bool  `json:"active"`
}

func (req *recurringRequest) validate(w http.ResponseWriter) bool {
	req.Description = strings.TrimSpace(req.Description)
	if req.Description == "" {
		writeError(w, http.StatusBadRequest, "a descrição é obrigatória")
		return false
	}
	if req.Frequency != "monthly" && req.Frequency != "annual" {
		writeError(w, http.StatusBadRequest, "frequency deve ser 'monthly' ou 'annual'")
		return false
	}
	if req.ExpectedAmountCents < 0 {
		writeError(w, http.StatusBadRequest, "o valor esperado não pode ser negativo")
		return false
	}
	if req.DueDay != nil && (*req.DueDay < 1 || *req.DueDay > 31) {
		writeError(w, http.StatusBadRequest, "dueDay deve estar entre 1 e 31")
		return false
	}
	if req.Frequency == "annual" && (req.DueMonth == nil || *req.DueMonth < 1 || *req.DueMonth > 12) {
		writeError(w, http.StatusBadRequest, "despesas anuais precisam de dueMonth (1–12)")
		return false
	}
	return true
}

func (s *Server) handleCreateRecurring(w http.ResponseWriter, r *http.Request) {
	var req recurringRequest
	if !readJSON(w, r, &req) || !req.validate(w) {
		return
	}
	rec, err := scanRecurring(s.Store.Pool.QueryRow(r.Context(),
		`INSERT INTO recurring_expenses (category_id, description, expected_amount_cents, frequency, due_day, due_month)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING `+recurringCols,
		req.CategoryID, req.Description, req.ExpectedAmountCents, req.Frequency, req.DueDay, req.DueMonth))
	if err != nil {
		writeError(w, http.StatusBadRequest, "não foi possível criar a recorrente (categoria existe?)")
		return
	}
	writeJSON(w, http.StatusCreated, rec)
}

func (s *Server) handleUpdateRecurring(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	var req recurringRequest
	if !readJSON(w, r, &req) || !req.validate(w) {
		return
	}
	active := true
	if req.Active != nil {
		active = *req.Active
	}
	rec, err := scanRecurring(s.Store.Pool.QueryRow(r.Context(),
		`UPDATE recurring_expenses
		 SET category_id = $2, description = $3, expected_amount_cents = $4,
		     frequency = $5, due_day = $6, due_month = $7, active = $8
		 WHERE id = $1 RETURNING `+recurringCols,
		id, req.CategoryID, req.Description, req.ExpectedAmountCents,
		req.Frequency, req.DueDay, req.DueMonth, active))
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "recorrente não encontrada")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "não foi possível atualizar a recorrente")
		return
	}
	writeJSON(w, http.StatusOK, rec)
}

func (s *Server) handleDeleteRecurring(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	// as previstas ainda por confirmar desta recorrente deixam de fazer sentido
	if _, err := s.Store.Pool.Exec(r.Context(),
		`DELETE FROM expenses WHERE recurring_id = $1 AND status = 'expected'`, id); err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	tag, err := s.Store.Pool.Exec(r.Context(), `DELETE FROM recurring_expenses WHERE id = $1`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "recorrente não encontrada")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ensureExpected materializa, de forma idempotente, as despesas previstas das
// recorrentes ativas para o mês dado (mensais sempre; anuais no mês devido).
// Não recua para antes do mês em que a recorrente foi criada.
func (s *Server) ensureExpected(ctx context.Context, month time.Time) error {
	daysInMonth := month.AddDate(0, 1, -1).Day()
	_, err := s.Store.Pool.Exec(ctx,
		`INSERT INTO expenses (category_id, description, amount_cents, spent_on, status, recurring_id)
		 SELECT r.category_id, r.description, r.expected_amount_cents,
		        make_date($2, $3, LEAST(COALESCE(r.due_day, 1), $4)), 'expected', r.id
		 FROM recurring_expenses r
		 WHERE r.active
		   AND date_trunc('month', r.created_at) <= $1
		   AND (r.frequency = 'monthly' OR (r.frequency = 'annual' AND r.due_month = $3))
		 ON CONFLICT (recurring_id, date_trunc('month', spent_on::timestamp))
		    WHERE recurring_id IS NOT NULL DO NOTHING`,
		month, month.Year(), int(month.Month()), daysInMonth)
	return err
}
