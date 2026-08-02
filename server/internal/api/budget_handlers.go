package api

import (
	"net/http"
	"time"
)

type Budget struct {
	CategoryID    int    `json:"categoryId"`
	Period        string `json:"period"`
	AmountCents   int64  `json:"amountCents"`
	Rollover      bool   `json:"rollover"`
	EffectiveFrom string `json:"effectiveFrom"`
}

// handleListBudgets devolve, para cada categoria e período, o orçamento em
// vigor no mês pedido (?month=YYYY-MM, mês atual por omissão).
func (s *Server) handleListBudgets(w http.ResponseWriter, r *http.Request) {
	month, ok := monthParam(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "mês inválido (esperado YYYY-MM)")
		return
	}
	rows, err := s.Store.Pool.Query(r.Context(),
		`SELECT DISTINCT ON (category_id, period)
		        category_id, period, amount_cents, rollover, effective_from
		 FROM budgets
		 WHERE effective_from <= $1
		 ORDER BY category_id, period, effective_from DESC`, month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	defer rows.Close()
	budgets := []Budget{}
	for rows.Next() {
		var b Budget
		var from time.Time
		if err := rows.Scan(&b.CategoryID, &b.Period, &b.AmountCents, &b.Rollover, &from); err != nil {
			writeError(w, http.StatusInternalServerError, "erro interno")
			return
		}
		b.EffectiveFrom = from.Format("2006-01")
		budgets = append(budgets, b)
	}
	writeJSON(w, http.StatusOK, budgets)
}

type budgetRequest struct {
	CategoryID  int    `json:"categoryId"`
	Period      string `json:"period"`
	AmountCents int64  `json:"amountCents"`
	Rollover    bool   `json:"rollover"`
	Month       string `json:"month"`
}

// handleSetBudget define o orçamento de uma categoria a partir de um mês
// (upsert em category+period+mês). amountCents = 0 remove o orçamento daí
// em diante.
func (s *Server) handleSetBudget(w http.ResponseWriter, r *http.Request) {
	var req budgetRequest
	if !readJSON(w, r, &req) {
		return
	}
	if req.Period != "monthly" && req.Period != "annual" {
		writeError(w, http.StatusBadRequest, "period deve ser 'monthly' ou 'annual'")
		return
	}
	if req.AmountCents < 0 {
		writeError(w, http.StatusBadRequest, "o valor não pode ser negativo")
		return
	}
	month, ok := parseMonth(req.Month)
	if !ok {
		writeError(w, http.StatusBadRequest, "mês inválido (esperado YYYY-MM)")
		return
	}
	_, err := s.Store.Pool.Exec(r.Context(),
		`INSERT INTO budgets (category_id, period, amount_cents, rollover, effective_from)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (category_id, period, effective_from)
		 DO UPDATE SET amount_cents = EXCLUDED.amount_cents, rollover = EXCLUDED.rollover`,
		req.CategoryID, req.Period, req.AmountCents, req.Rollover, month)
	if err != nil {
		writeError(w, http.StatusBadRequest, "não foi possível gravar o orçamento (categoria existe?)")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
