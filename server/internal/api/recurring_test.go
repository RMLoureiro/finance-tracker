package api_test

import (
	"fmt"
	"testing"
	"time"
)

func TestRecurringGeneratesAndConfirmsExpected(t *testing.T) {
	c := newTestClient(t)
	setupUser(c)
	cat := c.createCategory("Água", "#4A8AC9")

	// mensal (dia 12) e anual (setembro)
	resp, out := c.do("POST", "/api/recurring", map[string]any{
		"categoryId": cat, "description": "Fatura da água",
		"expectedAmountCents": 2800, "frequency": "monthly", "dueDay": 12})
	c.mustStatus(resp, 201, out)
	monthlyID := int(out["id"].(float64))

	resp, out = c.do("POST", "/api/recurring", map[string]any{
		"categoryId": cat, "description": "Seguro do carro",
		"expectedAmountCents": 18450, "frequency": "annual", "dueMonth": 9})
	c.mustStatus(resp, 201, out)

	// anual sem dueMonth é rejeitada
	resp, out = c.do("POST", "/api/recurring", map[string]any{
		"categoryId": cat, "description": "Inválida",
		"expectedAmountCents": 100, "frequency": "annual"})
	c.mustStatus(resp, 400, out)

	// as recorrentes foram criadas este mês, por isso as previstas nascem no
	// mês corrente (não recuam no tempo)
	now := time.Now()
	thisMonth := fmt.Sprintf("%04d-%02d", now.Year(), int(now.Month()))
	resp, out = c.do("GET", "/api/expenses?month="+thisMonth, nil)
	c.mustStatus(resp, 200, out)

	resp, list := c.doList("GET", "/api/expenses?month="+thisMonth)
	c.mustStatus(resp, 200, nil)
	var expected map[string]any
	for _, raw := range list {
		e := raw.(map[string]any)
		if e["status"] == "expected" && int(e["recurringId"].(float64)) == monthlyID {
			expected = e
		}
	}
	if expected == nil {
		t.Fatalf("prevista mensal não foi gerada: %v", list)
	}
	if expected["amountCents"].(float64) != 2800 {
		t.Errorf("valor previsto errado: %v", expected)
	}

	// pedir o mês outra vez não duplica a prevista
	c.do("GET", "/api/months/"+thisMonth, nil)
	resp, list = c.doList("GET", "/api/expenses?month="+thisMonth)
	c.mustStatus(resp, 200, nil)
	n := 0
	for _, raw := range list {
		if raw.(map[string]any)["status"] == "expected" {
			n++
		}
	}
	wantExpected := 1
	if now.Month() == time.September {
		wantExpected = 2 // a anual também cai neste mês
	}
	if n != wantExpected {
		t.Fatalf("esperava %d previstas, veio %d", wantExpected, n)
	}

	// confirmar a prevista com o valor real passa a gasto
	id := int(expected["id"].(float64))
	resp, out = c.do("PUT", fmt.Sprintf("/api/expenses/%d", id), map[string]any{
		"categoryId": cat, "description": "Fatura da água",
		"amountCents": 2650, "spentOn": expected["spentOn"], "status": "confirmed"})
	c.mustStatus(resp, 200, out)

	resp, out = c.do("GET", "/api/months/"+thisMonth, nil)
	c.mustStatus(resp, 200, out)
	totals := out["totals"].(map[string]any)
	if totals["spentCents"].(float64) != 2650 {
		t.Errorf("gasto = %v, esperado 2650", totals["spentCents"])
	}
	if wantExpected == 1 && totals["expectedCents"].(float64) != 0 {
		t.Errorf("previstas por confirmar = %v, esperado 0", totals["expectedCents"])
	}
	if out["projectedCents"] == nil {
		t.Error("projeção em falta no mês corrente")
	}

	// desativar a recorrente e apagar: previstas futuras deixam de ser geradas
	resp, out = c.do("DELETE", fmt.Sprintf("/api/recurring/%d", monthlyID), nil)
	c.mustStatus(resp, 200, out)
}

func TestRolloverAccumulatesAcrossMonths(t *testing.T) {
	c := newTestClient(t)
	setupUser(c)
	cat := c.createCategory("Comida dos animais", "#0F9488")

	// 100 €/mês com rollover desde junho de 2026
	resp, out := c.do("PUT", "/api/budgets", map[string]any{
		"categoryId": cat, "period": "monthly", "amountCents": 10000,
		"rollover": true, "month": "2026-06"})
	c.mustStatus(resp, 200, out)

	// junho: gasta 80 € (+20), julho: gasta 90 € (+10) → agosto começa com +30
	for _, e := range []map[string]any{
		{"categoryId": cat, "amountCents": 8000, "spentOn": "2026-06-15"},
		{"categoryId": cat, "amountCents": 9000, "spentOn": "2026-07-15"},
	} {
		resp, out := c.do("POST", "/api/expenses", e)
		c.mustStatus(resp, 201, out)
	}

	resp, out = c.do("GET", "/api/months/2026-08", nil)
	c.mustStatus(resp, 200, out)
	cats := out["categories"].([]any)
	if len(cats) != 1 {
		t.Fatalf("esperava 1 categoria, veio %d", len(cats))
	}
	cat0 := cats[0].(map[string]any)
	if cat0["rolloverCents"].(float64) != 3000 {
		t.Errorf("rollover = %v, esperado 3000", cat0["rolloverCents"])
	}
	if cat0["availableCents"].(float64) != 13000 {
		t.Errorf("disponível = %v, esperado 13000 (10000 + 3000)", cat0["availableCents"])
	}
}

func TestMonthlyAndAnnualReports(t *testing.T) {
	c := newTestClient(t)
	setupUser(c)
	agua := c.createCategory("Água", "#4A8AC9")
	gas := c.createCategory("Gás", "#C25E37")

	resp, out := c.do("PUT", "/api/budgets", map[string]any{
		"categoryId": agua, "period": "monthly", "amountCents": 3000, "month": "2026-01"})
	c.mustStatus(resp, 200, out)

	for _, e := range []map[string]any{
		{"categoryId": agua, "amountCents": 2500, "spentOn": "2026-06-10"},
		{"categoryId": agua, "amountCents": 2700, "spentOn": "2026-07-10"},
		{"categoryId": gas, "amountCents": 3100, "spentOn": "2026-07-18"},
	} {
		resp, out := c.do("POST", "/api/expenses", e)
		c.mustStatus(resp, 201, out)
	}

	resp, list := c.doList("GET", "/api/reports/monthly?end=2026-08&months=3")
	c.mustStatus(resp, 200, nil)
	if len(list) != 3 {
		t.Fatalf("esperava 3 meses, veio %d", len(list))
	}
	jul := list[1].(map[string]any)
	if jul["month"] != "2026-07" || jul["spentCents"].(float64) != 5800 {
		t.Errorf("julho errado: %v", jul)
	}
	if jul["budgetCents"].(float64) != 3000 {
		t.Errorf("orçamento de julho = %v, esperado 3000", jul["budgetCents"])
	}

	resp, list = c.doList("GET", "/api/reports/annual?year=2026")
	c.mustStatus(resp, 200, nil)
	for _, raw := range list {
		row := raw.(map[string]any)
		if int(row["categoryId"].(float64)) == agua {
			if row["spentCents"].(float64) != 5200 {
				t.Errorf("gasto anual da Água = %v, esperado 5200", row["spentCents"])
			}
			// orçamento anualizado: 3000 × 12
			if row["budgetCents"].(float64) != 36000 {
				t.Errorf("orçamento anual da Água = %v, esperado 36000", row["budgetCents"])
			}
		}
	}
}
