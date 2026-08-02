package api

import (
	"net/http"
	"time"
)

type MonthCategory struct {
	CategoryID         int   `json:"categoryId"`
	MonthlyBudgetCents int64 `json:"monthlyBudgetCents"`
	AnnualBudgetCents  int64 `json:"annualBudgetCents"`
	DilutedCents       int64 `json:"dilutedCents"`
	BudgetCents        int64 `json:"budgetCents"`
	Rollover           bool  `json:"rollover"`
	RolloverCents      int64 `json:"rolloverCents"`
	SpentCents         int64 `json:"spentCents"`
	ExpectedCents      int64 `json:"expectedCents"`
	AvailableCents     int64 `json:"availableCents"`
}

type MonthTotals struct {
	BudgetCents    int64 `json:"budgetCents"`
	SpentCents     int64 `json:"spentCents"`
	ExpectedCents  int64 `json:"expectedCents"`
	AvailableCents int64 `json:"availableCents"`
}

type MonthSummary struct {
	Month      string          `json:"month"`
	Totals     MonthTotals     `json:"totals"`
	Categories []MonthCategory `json:"categories"`
}

// monthSummary calcula o resumo de um mês: orçamento em vigor por categoria
// (mensal + anual diluído a 1/12), gasto confirmado e previstas por confirmar.
func (s *Server) monthSummary(r *http.Request, month time.Time) (*MonthSummary, error) {
	ctx := r.Context()
	next := month.AddDate(0, 1, 0)
	byCat := map[int]*MonthCategory{}
	order := []int{}

	cat := func(id int) *MonthCategory {
		if c, ok := byCat[id]; ok {
			return c
		}
		c := &MonthCategory{CategoryID: id}
		byCat[id] = c
		order = append(order, id)
		return c
	}

	rows, err := s.Store.Pool.Query(ctx,
		`SELECT DISTINCT ON (category_id, period) category_id, period, amount_cents, rollover
		 FROM budgets WHERE effective_from <= $1
		 ORDER BY category_id, period, effective_from DESC`, month)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		var period string
		var amount int64
		var rollover bool
		if err := rows.Scan(&id, &period, &amount, &rollover); err != nil {
			rows.Close()
			return nil, err
		}
		c := cat(id)
		if period == "monthly" {
			c.MonthlyBudgetCents = amount
			c.Rollover = rollover
		} else {
			c.AnnualBudgetCents = amount
			c.DilutedCents = amount / 12
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	rows, err = s.Store.Pool.Query(ctx,
		`SELECT category_id,
		        COALESCE(sum(amount_cents) FILTER (WHERE status = 'confirmed'), 0),
		        COALESCE(sum(amount_cents) FILTER (WHERE status = 'expected'), 0)
		 FROM expenses WHERE spent_on >= $1 AND spent_on < $2
		 GROUP BY category_id`, month, next)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id int
		var spent, expected int64
		if err := rows.Scan(&id, &spent, &expected); err != nil {
			rows.Close()
			return nil, err
		}
		c := cat(id)
		c.SpentCents = spent
		c.ExpectedCents = expected
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	sum := &MonthSummary{Month: month.Format("2006-01"), Categories: []MonthCategory{}}
	for _, id := range order {
		c := byCat[id]
		c.BudgetCents = c.MonthlyBudgetCents + c.DilutedCents
		c.AvailableCents = c.BudgetCents + c.RolloverCents - c.SpentCents
		sum.Totals.BudgetCents += c.BudgetCents
		sum.Totals.SpentCents += c.SpentCents
		sum.Totals.ExpectedCents += c.ExpectedCents
		sum.Categories = append(sum.Categories, *c)
	}
	sum.Totals.AvailableCents = sum.Totals.BudgetCents - sum.Totals.SpentCents
	return sum, nil
}

func (s *Server) handleMonthSummary(w http.ResponseWriter, r *http.Request) {
	month, ok := monthParam(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "mês inválido (esperado YYYY-MM)")
		return
	}
	sum, err := s.monthSummary(r, month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, sum)
}
