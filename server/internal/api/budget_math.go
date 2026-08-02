package api

import (
	"context"
	"time"
)

type budgetVersion struct {
	from     time.Time
	amount   int64
	rollover bool
}

// budgetHistory guarda todas as versões de orçamento por categoria e período,
// ordenadas por effective_from ascendente.
type budgetHistory map[int]map[string][]budgetVersion

func (s *Server) loadBudgetHistory(ctx context.Context) (budgetHistory, error) {
	rows, err := s.Store.Pool.Query(ctx,
		`SELECT category_id, period, amount_cents, rollover, effective_from
		 FROM budgets ORDER BY effective_from`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	h := budgetHistory{}
	for rows.Next() {
		var catID int
		var period string
		var v budgetVersion
		if err := rows.Scan(&catID, &period, &v.amount, &v.rollover, &v.from); err != nil {
			return nil, err
		}
		if h[catID] == nil {
			h[catID] = map[string][]budgetVersion{}
		}
		h[catID][period] = append(h[catID][period], v)
	}
	return h, rows.Err()
}

// at devolve a versão em vigor num mês (a última com effective_from <= mês).
func (h budgetHistory) at(catID int, period string, month time.Time) (int64, bool, bool) {
	var out budgetVersion
	found := false
	for _, v := range h[catID][period] {
		if v.from.After(month) {
			break
		}
		out = v
		found = true
	}
	return out.amount, out.rollover, found
}

// monthlyBudget devolve o orçamento efetivo de uma categoria num mês:
// mensal + anual diluído a 1/12.
func (h budgetHistory) monthlyBudget(catID int, month time.Time) (monthly, annual, diluted int64, rollover bool) {
	monthly, rollover, _ = h.at(catID, "monthly", month)
	annual, _, _ = h.at(catID, "annual", month)
	diluted = annual / 12
	return
}

// firstMonth devolve o primeiro mês com orçamento definido para a categoria.
func (h budgetHistory) firstMonth(catID int) (time.Time, bool) {
	var first time.Time
	found := false
	for _, versions := range h[catID] {
		for _, v := range versions {
			if !found || v.from.Before(first) {
				first = v.from
				found = true
			}
		}
	}
	return first, found
}

func (h budgetHistory) categoryIDs() []int {
	ids := make([]int, 0, len(h))
	for id := range h {
		ids = append(ids, id)
	}
	return ids
}
