package api_test

import (
	"fmt"
	"testing"
)

func setupUser(c *client) {
	resp, out := c.do("POST", "/api/setup", map[string]string{"password": "password-segura"})
	c.mustStatus(resp, 201, out)
}

func (c *client) createCategory(name, color string) int {
	c.t.Helper()
	resp, out := c.do("POST", "/api/categories", map[string]any{"name": name, "color": color, "icon": "•"})
	c.mustStatus(resp, 201, out)
	return int(out["id"].(float64))
}

func TestBudgetsExpensesAndMonthSummary(t *testing.T) {
	c := newTestClient(t)
	setupUser(c)

	agua := c.createCategory("Água", "#4A8AC9")
	gasolina := c.createCategory("Gasolina", "#8A64B8")

	// orçamentos: Água 30 €/mês + 24 €/ano (diluído: 2 €/mês); Gasolina 120 €/mês
	for _, b := range []map[string]any{
		{"categoryId": agua, "period": "monthly", "amountCents": 3000, "month": "2026-01"},
		{"categoryId": agua, "period": "annual", "amountCents": 2400, "month": "2026-01"},
		{"categoryId": gasolina, "period": "monthly", "amountCents": 12000, "month": "2026-01"},
	} {
		resp, out := c.do("PUT", "/api/budgets", b)
		c.mustStatus(resp, 200, out)
	}

	// despesas de agosto: confirmadas e uma prevista
	for _, e := range []map[string]any{
		{"categoryId": agua, "amountCents": 2200, "spentOn": "2026-08-12", "description": "Fatura da água"},
		{"categoryId": gasolina, "amountCents": 4520, "spentOn": "2026-08-02", "description": "BP"},
		{"categoryId": agua, "amountCents": 2800, "spentOn": "2026-08-20", "status": "expected"},
	} {
		resp, out := c.do("POST", "/api/expenses", e)
		c.mustStatus(resp, 201, out)
	}

	resp, out := c.do("GET", "/api/months/2026-08", nil)
	c.mustStatus(resp, 200, out)
	totals := out["totals"].(map[string]any)
	if got := totals["budgetCents"].(float64); got != 15200 {
		t.Errorf("budget total = %v, esperado 15200", got)
	}
	if got := totals["spentCents"].(float64); got != 6720 {
		t.Errorf("spent total = %v, esperado 6720", got)
	}
	if got := totals["expectedCents"].(float64); got != 2800 {
		t.Errorf("expected total = %v, esperado 2800", got)
	}
	if got := totals["availableCents"].(float64); got != 8480 {
		t.Errorf("available total = %v, esperado 8480", got)
	}
	cats := out["categories"].([]any)
	if len(cats) != 2 {
		t.Fatalf("esperava 2 categorias no resumo, veio %d", len(cats))
	}
	for _, raw := range cats {
		cat := raw.(map[string]any)
		if int(cat["categoryId"].(float64)) == agua {
			if cat["budgetCents"].(float64) != 3200 || cat["dilutedCents"].(float64) != 200 {
				t.Errorf("resumo da Água errado: %v", cat)
			}
			if cat["availableCents"].(float64) != 1000 {
				t.Errorf("disponível da Água = %v, esperado 1000", cat["availableCents"])
			}
		}
	}

	// alterar o orçamento a partir de setembro não muda agosto
	resp, out = c.do("PUT", "/api/budgets",
		map[string]any{"categoryId": agua, "period": "monthly", "amountCents": 5000, "month": "2026-09"})
	c.mustStatus(resp, 200, out)
	resp, out = c.do("GET", "/api/months/2026-08", nil)
	c.mustStatus(resp, 200, out)
	if got := out["totals"].(map[string]any)["budgetCents"].(float64); got != 15200 {
		t.Errorf("budget de agosto mudou para %v depois de alterar setembro", got)
	}
	resp, out = c.do("GET", "/api/months/2026-09", nil)
	c.mustStatus(resp, 200, out)
	if got := out["totals"].(map[string]any)["budgetCents"].(float64); got != 17200 {
		t.Errorf("budget de setembro = %v, esperado 17200", got)
	}
}

func TestExpenseCRUDAndCategoryDelete(t *testing.T) {
	c := newTestClient(t)
	setupUser(c)

	cat := c.createCategory("Supermercado", "#C94F72")
	vazia := c.createCategory("Sem uso", "#0F9488")

	resp, out := c.do("POST", "/api/expenses",
		map[string]any{"categoryId": cat, "amountCents": 6235, "spentOn": "2026-08-02", "description": "Continente"})
	c.mustStatus(resp, 201, out)
	id := int(out["id"].(float64))

	// editar
	resp, out = c.do("PUT", fmt.Sprintf("/api/expenses/%d", id),
		map[string]any{"categoryId": cat, "amountCents": 6000, "spentOn": "2026-08-03", "description": "Continente (corrigido)"})
	c.mustStatus(resp, 200, out)
	if out["amountCents"].(float64) != 6000 || out["spentOn"] != "2026-08-03" {
		t.Errorf("edição não aplicada: %v", out)
	}

	// listar por mês
	resp, _ = c.do("GET", "/api/expenses?month=2026-08", nil)
	c.mustStatus(resp, 200, nil)

	// valores inválidos
	resp, out = c.do("POST", "/api/expenses",
		map[string]any{"categoryId": cat, "amountCents": 0, "spentOn": "2026-08-02"})
	c.mustStatus(resp, 400, out)
	resp, out = c.do("POST", "/api/expenses",
		map[string]any{"categoryId": 99999, "amountCents": 100, "spentOn": "2026-08-02"})
	c.mustStatus(resp, 400, out)

	// apagar categoria com despesas → conflito; sem despesas → ok
	resp, out = c.do("DELETE", fmt.Sprintf("/api/categories/%d", cat), nil)
	c.mustStatus(resp, 409, out)
	resp, out = c.do("DELETE", fmt.Sprintf("/api/categories/%d", vazia), nil)
	c.mustStatus(resp, 200, out)

	// apagar despesa
	resp, out = c.do("DELETE", fmt.Sprintf("/api/expenses/%d", id), nil)
	c.mustStatus(resp, 200, out)
	resp, out = c.do("DELETE", fmt.Sprintf("/api/expenses/%d", id), nil)
	c.mustStatus(resp, 404, out)
}
