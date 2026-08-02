package api

import (
	"net/http"
	"sort"
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
	RolloverCents  int64 `json:"rolloverCents"`
	SpentCents     int64 `json:"spentCents"`
	ExpectedCents  int64 `json:"expectedCents"`
	AvailableCents int64 `json:"availableCents"`
}

type MonthSummary struct {
	Month          string          `json:"month"`
	DaysInMonth    int             `json:"daysInMonth"`
	DayOfMonth     *int            `json:"dayOfMonth"`     // só no mês corrente
	ProjectedCents *int64          `json:"projectedCents"` // só no mês corrente
	Totals         MonthTotals     `json:"totals"`
	Categories     []MonthCategory `json:"categories"`
}

type monthKey struct {
	year  int
	month time.Month
}

func keyOf(t time.Time) monthKey { return monthKey{t.Year(), t.Month()} }

// monthSummary calcula o resumo de um mês: orçamento em vigor por categoria
// (mensal + anual diluído a 1/12), rollover acumulado, gasto confirmado,
// previstas por confirmar e, no mês corrente, a projeção de fim de mês.
func (s *Server) monthSummary(r *http.Request, month time.Time) (*MonthSummary, error) {
	ctx := r.Context()
	if err := s.ensureExpected(ctx, month); err != nil {
		return nil, err
	}
	hist, err := s.loadBudgetHistory(ctx)
	if err != nil {
		return nil, err
	}
	next := month.AddDate(0, 1, 0)

	byCat := map[int]*MonthCategory{}
	cat := func(id int) *MonthCategory {
		if c, ok := byCat[id]; ok {
			return c
		}
		c := &MonthCategory{CategoryID: id}
		byCat[id] = c
		return c
	}

	for _, id := range hist.categoryIDs() {
		monthly, annual, diluted, rollover := hist.monthlyBudget(id, month)
		if monthly == 0 && annual == 0 {
			continue
		}
		c := cat(id)
		c.MonthlyBudgetCents = monthly
		c.AnnualBudgetCents = annual
		c.DilutedCents = diluted
		c.Rollover = rollover
	}

	rows, err := s.Store.Pool.Query(ctx,
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

	// rollover: acumula (orçamento − gasto) de cada mês anterior, desde o
	// primeiro mês com orçamento, para as categorias com rollover ativo
	needsRollover := false
	for _, c := range byCat {
		if c.Rollover {
			needsRollover = true
		}
	}
	if needsRollover {
		spentBefore := map[int]map[monthKey]int64{}
		rows, err = s.Store.Pool.Query(ctx,
			`SELECT category_id, date_trunc('month', spent_on)::date, sum(amount_cents)
			 FROM expenses WHERE status = 'confirmed' AND spent_on < $1
			 GROUP BY 1, 2`, month)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var id int
			var m time.Time
			var v int64
			if err := rows.Scan(&id, &m, &v); err != nil {
				rows.Close()
				return nil, err
			}
			if spentBefore[id] == nil {
				spentBefore[id] = map[monthKey]int64{}
			}
			spentBefore[id][keyOf(m)] = v
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}

		for id, c := range byCat {
			if !c.Rollover {
				continue
			}
			first, ok := hist.firstMonth(id)
			if !ok {
				continue
			}
			var roll int64
			for m := first; m.Before(month); m = m.AddDate(0, 1, 0) {
				monthly, _, diluted, _ := hist.monthlyBudget(id, m)
				roll += monthly + diluted - spentBefore[id][keyOf(m)]
			}
			c.RolloverCents = roll
		}
	}

	sum := &MonthSummary{
		Month:       month.Format("2006-01"),
		DaysInMonth: month.AddDate(0, 1, -1).Day(),
		Categories:  []MonthCategory{},
	}
	ids := make([]int, 0, len(byCat))
	for id := range byCat {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	for _, id := range ids {
		c := byCat[id]
		c.BudgetCents = c.MonthlyBudgetCents + c.DilutedCents
		c.AvailableCents = c.BudgetCents + c.RolloverCents - c.SpentCents
		sum.Totals.BudgetCents += c.BudgetCents
		sum.Totals.RolloverCents += c.RolloverCents
		sum.Totals.SpentCents += c.SpentCents
		sum.Totals.ExpectedCents += c.ExpectedCents
		sum.Categories = append(sum.Categories, *c)
	}
	sum.Totals.AvailableCents = sum.Totals.BudgetCents + sum.Totals.RolloverCents - sum.Totals.SpentCents

	// projeção: só faz sentido para o mês em curso
	now := time.Now()
	if keyOf(now) == keyOf(month) {
		day := now.Day()
		sum.DayOfMonth = &day

		var confirmedRecurring int64
		err := s.Store.Pool.QueryRow(ctx,
			`SELECT COALESCE(sum(amount_cents), 0) FROM expenses
			 WHERE spent_on >= $1 AND spent_on < $2 AND status = 'confirmed' AND recurring_id IS NOT NULL`,
			month, next).Scan(&confirmedRecurring)
		if err != nil {
			return nil, err
		}
		// fixas já pagas + fixas por pagar + gasto variável extrapolado ao ritmo atual
		variable := sum.Totals.SpentCents - confirmedRecurring
		projected := confirmedRecurring + sum.Totals.ExpectedCents +
			int64(float64(variable)/float64(day)*float64(sum.DaysInMonth))
		sum.ProjectedCents = &projected
	}
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
