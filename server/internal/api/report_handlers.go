package api

import (
	"net/http"
	"sort"
	"strconv"
	"time"
)

type ReportMonth struct {
	Month       string           `json:"month"`
	SpentCents  int64            `json:"spentCents"`
	BudgetCents int64            `json:"budgetCents"`
	ByCategory  map[string]int64 `json:"byCategory"`
}

// handleMonthlyReport devolve o gasto confirmado e o orçamento em vigor dos
// últimos N meses (?months=, por omissão 12, máx. 36) a terminar em ?end=YYYY-MM.
func (s *Server) handleMonthlyReport(w http.ResponseWriter, r *http.Request) {
	end := time.Now()
	end = time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, time.UTC)
	if v := r.URL.Query().Get("end"); v != "" {
		var ok bool
		if end, ok = parseMonth(v); !ok {
			writeError(w, http.StatusBadRequest, "end inválido (esperado YYYY-MM)")
			return
		}
	}
	months := 12
	if v := r.URL.Query().Get("months"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 36 {
			writeError(w, http.StatusBadRequest, "months deve estar entre 1 e 36")
			return
		}
		months = n
	}
	start := end.AddDate(0, -(months - 1), 0)

	hist, err := s.loadBudgetHistory(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	spent := map[monthKey]map[int]int64{}
	rows, err := s.Store.Pool.Query(r.Context(),
		`SELECT date_trunc('month', spent_on)::date, category_id, sum(amount_cents)
		 FROM expenses
		 WHERE status = 'confirmed' AND spent_on >= $1 AND spent_on < $2
		 GROUP BY 1, 2`, start, end.AddDate(0, 1, 0))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	for rows.Next() {
		var m time.Time
		var catID int
		var v int64
		if err := rows.Scan(&m, &catID, &v); err != nil {
			rows.Close()
			writeError(w, http.StatusInternalServerError, "erro interno")
			return
		}
		if spent[keyOf(m)] == nil {
			spent[keyOf(m)] = map[int]int64{}
		}
		spent[keyOf(m)][catID] = v
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	out := []ReportMonth{}
	for m := start; !m.After(end); m = m.AddDate(0, 1, 0) {
		rm := ReportMonth{Month: m.Format("2006-01"), ByCategory: map[string]int64{}}
		for catID, v := range spent[keyOf(m)] {
			rm.ByCategory[strconv.Itoa(catID)] = v
			rm.SpentCents += v
		}
		for _, catID := range hist.categoryIDs() {
			monthly, _, diluted, _ := hist.monthlyBudget(catID, m)
			rm.BudgetCents += monthly + diluted
		}
		out = append(out, rm)
	}
	writeJSON(w, http.StatusOK, out)
}

type AnnualCategoryReport struct {
	CategoryID  int   `json:"categoryId"`
	SpentCents  int64 `json:"spentCents"`
	BudgetCents int64 `json:"budgetCents"`
}

// handleAnnualReport devolve, por categoria, o gasto do ano e o orçamento
// anualizado (soma dos orçamentos mensais efetivos do ano).
func (s *Server) handleAnnualReport(w http.ResponseWriter, r *http.Request) {
	year := time.Now().Year()
	if v := r.URL.Query().Get("year"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 2000 || n > 2200 {
			writeError(w, http.StatusBadRequest, "year inválido")
			return
		}
		year = n
	}
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)

	hist, err := s.loadBudgetHistory(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	byCat := map[int]*AnnualCategoryReport{}
	get := func(id int) *AnnualCategoryReport {
		if c, ok := byCat[id]; ok {
			return c
		}
		c := &AnnualCategoryReport{CategoryID: id}
		byCat[id] = c
		return c
	}

	rows, err := s.Store.Pool.Query(r.Context(),
		`SELECT category_id, sum(amount_cents)
		 FROM expenses
		 WHERE status = 'confirmed' AND spent_on >= $1 AND spent_on < $2
		 GROUP BY category_id`, start, start.AddDate(1, 0, 0))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	for rows.Next() {
		var id int
		var v int64
		if err := rows.Scan(&id, &v); err != nil {
			rows.Close()
			writeError(w, http.StatusInternalServerError, "erro interno")
			return
		}
		get(id).SpentCents = v
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	for _, catID := range hist.categoryIDs() {
		var budget int64
		for m := start; m.Year() == year; m = m.AddDate(0, 1, 0) {
			monthly, _, diluted, _ := hist.monthlyBudget(catID, m)
			budget += monthly + diluted
		}
		if budget > 0 {
			get(catID).BudgetCents = budget
		}
	}

	out := []AnnualCategoryReport{}
	for _, c := range byCat {
		out = append(out, *c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CategoryID < out[j].CategoryID })
	writeJSON(w, http.StatusOK, out)
}
