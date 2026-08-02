package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type Expense struct {
	ID          int    `json:"id"`
	CategoryID  int    `json:"categoryId"`
	Description string `json:"description"`
	AmountCents int64  `json:"amountCents"`
	SpentOn     string `json:"spentOn"`
	Status      string `json:"status"`
	RecurringID *int   `json:"recurringId"`
}

func scanExpense(row pgx.Row) (Expense, error) {
	var e Expense
	var spentOn time.Time
	err := row.Scan(&e.ID, &e.CategoryID, &e.Description, &e.AmountCents, &spentOn, &e.Status, &e.RecurringID)
	e.SpentOn = spentOn.Format("2006-01-02")
	return e, err
}

const expenseCols = `id, category_id, description, amount_cents, spent_on, status, recurring_id`

// handleListExpenses lista as despesas de um mês (?month=YYYY-MM), opcionalmente
// filtradas por categoria (?categoryId=).
func (s *Server) handleListExpenses(w http.ResponseWriter, r *http.Request) {
	month, ok := monthParam(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "mês inválido (esperado YYYY-MM)")
		return
	}
	next := month.AddDate(0, 1, 0)

	query := `SELECT ` + expenseCols + ` FROM expenses WHERE spent_on >= $1 AND spent_on < $2`
	args := []any{month, next}
	if cid := r.URL.Query().Get("categoryId"); cid != "" {
		id, err := strconv.Atoi(cid)
		if err != nil {
			writeError(w, http.StatusBadRequest, "categoryId inválido")
			return
		}
		query += ` AND category_id = $3`
		args = append(args, id)
	}
	query += ` ORDER BY spent_on DESC, id DESC`

	rows, err := s.Store.Pool.Query(r.Context(), query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	defer rows.Close()
	expenses := []Expense{}
	for rows.Next() {
		e, err := scanExpense(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "erro interno")
			return
		}
		expenses = append(expenses, e)
	}
	writeJSON(w, http.StatusOK, expenses)
}

type expenseRequest struct {
	CategoryID  int    `json:"categoryId"`
	Description string `json:"description"`
	AmountCents int64  `json:"amountCents"`
	SpentOn     string `json:"spentOn"`
	Status      string `json:"status"`
}

func (req *expenseRequest) validate(w http.ResponseWriter) (time.Time, bool) {
	req.Description = strings.TrimSpace(req.Description)
	if req.Status == "" {
		req.Status = "confirmed"
	}
	if req.Status != "expected" && req.Status != "confirmed" && req.Status != "skipped" {
		writeError(w, http.StatusBadRequest, "status inválido")
		return time.Time{}, false
	}
	if req.AmountCents < 0 || (req.Status == "confirmed" && req.AmountCents == 0) {
		writeError(w, http.StatusBadRequest, "o valor da despesa deve ser positivo")
		return time.Time{}, false
	}
	spentOn, err := time.Parse("2006-01-02", req.SpentOn)
	if err != nil {
		writeError(w, http.StatusBadRequest, "data inválida (esperado YYYY-MM-DD)")
		return time.Time{}, false
	}
	return spentOn, true
}

func (s *Server) handleCreateExpense(w http.ResponseWriter, r *http.Request) {
	var req expenseRequest
	if !readJSON(w, r, &req) {
		return
	}
	spentOn, ok := req.validate(w)
	if !ok {
		return
	}
	e, err := scanExpense(s.Store.Pool.QueryRow(r.Context(),
		`INSERT INTO expenses (category_id, description, amount_cents, spent_on, status)
		 VALUES ($1, $2, $3, $4, $5) RETURNING `+expenseCols,
		req.CategoryID, req.Description, req.AmountCents, spentOn, req.Status))
	if err != nil {
		writeError(w, http.StatusBadRequest, "não foi possível criar a despesa (categoria existe?)")
		return
	}
	writeJSON(w, http.StatusCreated, e)
}

func (s *Server) handleUpdateExpense(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	var req expenseRequest
	if !readJSON(w, r, &req) {
		return
	}
	spentOn, ok := req.validate(w)
	if !ok {
		return
	}
	e, err := scanExpense(s.Store.Pool.QueryRow(r.Context(),
		`UPDATE expenses
		 SET category_id = $2, description = $3, amount_cents = $4, spent_on = $5, status = $6
		 WHERE id = $1 RETURNING `+expenseCols,
		id, req.CategoryID, req.Description, req.AmountCents, spentOn, req.Status))
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "despesa não encontrada")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "não foi possível atualizar a despesa")
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (s *Server) handleDeleteExpense(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	tag, err := s.Store.Pool.Exec(r.Context(), `DELETE FROM expenses WHERE id = $1`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	if tag.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "despesa não encontrada")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
