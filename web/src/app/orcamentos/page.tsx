"use client";

import { useMemo, useState } from "react";
import { api, Budget, Category, Recurring } from "@/lib/api";
import { currentMonth, eurosCompact } from "@/lib/money";
import { useData } from "@/lib/hooks";
import BudgetSheet from "@/components/BudgetSheet";
import RecurringSheet from "@/components/RecurringSheet";
import CategorySheet from "@/components/CategorySheet";

const MONTH_SHORT = ["jan", "fev", "mar", "abr", "mai", "jun", "jul", "ago", "set", "out", "nov", "dez"];

type Sheet =
  | { kind: "none" }
  | { kind: "budget"; category: Category }
  | { kind: "recurring"; recurring: Recurring | null }
  | { kind: "category"; category: Category | null };

export default function OrcamentosPage() {
  const [sheet, setSheet] = useState<Sheet>({ kind: "none" });

  const { data: categories, reload: reloadCats } = useData(
    () => api.get<Category[]>("/api/categories"),
    []
  );
  const { data: budgets, reload: reloadBudgets } = useData(
    () => api.get<Budget[]>(`/api/budgets?month=${currentMonth()}`),
    []
  );
  const { data: recurring, reload: reloadRecurring } = useData(
    () => api.get<Recurring[]>("/api/recurring"),
    []
  );

  const catById = useMemo(() => {
    const m = new Map<number, Category>();
    (categories ?? []).forEach((c) => m.set(c.id, c));
    return m;
  }, [categories]);

  const budgetOf = (categoryId: number, period: "monthly" | "annual") =>
    (budgets ?? []).find(
      (b) => b.categoryId === categoryId && b.period === period && b.amountCents > 0
    ) ?? null;

  function closeAndReload() {
    setSheet({ kind: "none" });
    reloadCats();
    reloadBudgets();
    reloadRecurring();
  }

  async function logout() {
    await api.post("/api/logout", {});
    location.href = "/login";
  }

  return (
    <>
      <main className="content" style={{ paddingTop: 20 }}>
        <div className="cols-2">
          <section>
            <p className="section-label">Orçamentos por categoria</p>
            <div className="card">
              {(categories ?? [])
                .filter((c) => !c.archived)
                .map((c) => {
                  const monthly = budgetOf(c.id, "monthly");
                  const annual = budgetOf(c.id, "annual");
                  return (
                    <button
                      key={c.id}
                      className="row"
                      style={{ width: "100%", textAlign: "left" }}
                      onClick={() => setSheet({ kind: "budget", category: c })}
                    >
                      <span className="dot" style={{ background: c.color }} />
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div style={{ fontSize: 13.5, fontWeight: 600 }}>
                          {c.icon} {c.name}
                        </div>
                        <div style={{ fontSize: 11.5, color: "var(--muted)" }}>
                          {monthly || annual ? (
                            <>
                              {monthly && <>mensal {eurosCompact(monthly.amountCents)}</>}
                              {monthly?.rollover && " · rollover"}
                              {monthly && annual && " · "}
                              {annual && <>anual {eurosCompact(annual.amountCents)}</>}
                            </>
                          ) : (
                            "sem orçamento — toca para definir"
                          )}
                        </div>
                      </div>
                      <span style={{ color: "var(--muted)" }}>›</span>
                    </button>
                  );
                })}
            </div>
          </section>

          <section>
            <p className="section-label">Despesas fixas (recorrentes)</p>
            <div className="card">
              {(recurring ?? []).length === 0 && (
                <p style={{ color: "var(--muted)", fontSize: 13.5, padding: "12px 0" }}>
                  As despesas fixas (água, luz, seguros…) são criadas automaticamente
                  todos os meses como «previstas» — só confirmas o valor real.
                </p>
              )}
              {(recurring ?? []).map((r) => {
                const cat = catById.get(r.categoryId);
                return (
                  <button
                    key={r.id}
                    className="row"
                    style={{ width: "100%", textAlign: "left", opacity: r.active ? 1 : 0.55 }}
                    onClick={() => setSheet({ kind: "recurring", recurring: r })}
                  >
                    <span className="dot" style={{ background: cat?.color ?? "var(--muted)" }} />
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <div style={{ fontSize: 13.5, fontWeight: 600 }}>{r.description}</div>
                      <div style={{ fontSize: 11.5, color: "var(--muted)" }}>
                        {cat?.name} ·{" "}
                        {r.frequency === "monthly"
                          ? `mensal${r.dueDay ? `, dia ${r.dueDay}` : ""}`
                          : `anual, ${MONTH_SHORT[(r.dueMonth ?? 1) - 1]}`}
                        {!r.active && " · inativa"}
                      </div>
                    </div>
                    <span className="num" style={{ fontSize: 12.5, fontWeight: 700, color: "var(--muted)" }}>
                      ≈ {eurosCompact(r.expectedAmountCents)}
                    </span>
                  </button>
                );
              })}
              <div style={{ padding: "10px 0" }}>
                <button
                  className="btn-quiet"
                  style={{ padding: 0 }}
                  onClick={() => setSheet({ kind: "recurring", recurring: null })}
                >
                  ＋ Adicionar despesa fixa
                </button>
              </div>
            </div>

            <p className="section-label">Categorias</p>
            <div className="card">
              {(categories ?? []).map((c) => (
                <button
                  key={c.id}
                  className="row"
                  style={{ width: "100%", textAlign: "left", opacity: c.archived ? 0.55 : 1 }}
                  onClick={() => setSheet({ kind: "category", category: c })}
                >
                  <span className="dot" style={{ background: c.color }} />
                  <div style={{ flex: 1, fontSize: 13.5, fontWeight: 600 }}>
                    {c.icon} {c.name}
                    {c.archived && (
                      <span style={{ fontWeight: 400, color: "var(--muted)" }}> · arquivada</span>
                    )}
                  </div>
                  <span style={{ color: "var(--muted)" }}>›</span>
                </button>
              ))}
              <div style={{ padding: "10px 0" }}>
                <button
                  className="btn-quiet"
                  style={{ padding: 0 }}
                  onClick={() => setSheet({ kind: "category", category: null })}
                >
                  ＋ Nova categoria
                </button>
              </div>
            </div>

            <div style={{ textAlign: "center", marginTop: 24 }}>
              <button className="btn-danger" onClick={logout}>
                Terminar sessão
              </button>
            </div>
          </section>
        </div>
      </main>

      {sheet.kind === "budget" && (
        <BudgetSheet
          category={sheet.category}
          monthly={budgetOf(sheet.category.id, "monthly")}
          annual={budgetOf(sheet.category.id, "annual")}
          onClose={() => setSheet({ kind: "none" })}
          onSaved={closeAndReload}
        />
      )}
      {sheet.kind === "recurring" && categories && (
        <RecurringSheet
          categories={categories}
          initial={sheet.recurring}
          onClose={() => setSheet({ kind: "none" })}
          onSaved={closeAndReload}
        />
      )}
      {sheet.kind === "category" && (
        <CategorySheet
          initial={sheet.category}
          onClose={() => setSheet({ kind: "none" })}
          onSaved={closeAndReload}
        />
      )}
    </>
  );
}
