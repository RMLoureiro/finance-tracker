"use client";

import { useMemo, useState } from "react";
import { api, Category, Expense, MonthSummary } from "@/lib/api";
import { currentMonth, euros, eurosCompact } from "@/lib/money";
import { useData } from "@/lib/hooks";
import MonthNav from "@/components/MonthNav";
import ExpenseSheet from "@/components/ExpenseSheet";

export default function HomePage() {
  const [month, setMonth] = useState(currentMonth());
  const [sheet, setSheet] = useState<{ open: boolean; expense: Expense | null }>({
    open: false,
    expense: null,
  });

  const { data: categories } = useData(
    () => api.get<Category[]>("/api/categories"),
    []
  );
  const {
    data: summary,
    reload: reloadSummary,
  } = useData(() => api.get<MonthSummary>(`/api/months/${month}`), [month]);
  const { data: expenses, reload: reloadExpenses } = useData(
    () => api.get<Expense[]>(`/api/expenses?month=${month}`),
    [month]
  );

  const catById = useMemo(() => {
    const m = new Map<number, Category>();
    (categories ?? []).forEach((c) => m.set(c.id, c));
    return m;
  }, [categories]);

  const expected = (expenses ?? []).filter((e) => e.status === "expected");
  const recent = (expenses ?? []).filter((e) => e.status === "confirmed").slice(0, 5);
  const totals = summary?.totals;
  const isCurrent = month === currentMonth();

  function refresh() {
    setSheet({ open: false, expense: null });
    reloadSummary();
    reloadExpenses();
  }

  return (
    <>
      <header className="head">
        <MonthNav month={month} onChange={setMonth} />
        <div style={{ marginTop: 10, fontSize: 11, textTransform: "uppercase", letterSpacing: "0.1em", opacity: 0.65 }}>
          Disponível
        </div>
        <div className="avail num">
          {totals ? euros(totals.availableCents) : "—"}
        </div>
        {totals && (
          <div className="sub num">
            de {euros(totals.budgetCents + totals.rolloverCents)} orçamentados
            {totals.rolloverCents !== 0 && ` (inclui ${euros(totals.rolloverCents)} transitados)`}
            {isCurrent && summary?.projectedCents != null && (
              <> · projeção do mês: {euros(summary.projectedCents)}</>
            )}
          </div>
        )}
        {summary && totals && totals.budgetCents + totals.rolloverCents > 0 && (
          <div className="seg" aria-hidden>
            {summary.categories
              .filter((c) => c.spentCents > 0)
              .map((c) => (
                <i
                  key={c.categoryId}
                  style={{
                    flex: c.spentCents,
                    background: catById.get(c.categoryId)?.color ?? "var(--muted)",
                  }}
                />
              ))}
            <i
              style={{
                flex: Math.max(totals.budgetCents + totals.rolloverCents - totals.spentCents, 0),
                background: "rgba(255,255,255,0.18)",
              }}
            />
          </div>
        )}
      </header>

      <main className="content">
        <div className="cols-2">
          <section>
            <p className="section-label">Orçamentos</p>
            <div className="card">
              {!summary || summary.categories.length === 0 ? (
                <p style={{ color: "var(--muted)", fontSize: 13.5, padding: "12px 0" }}>
                  Ainda sem orçamentos. Define-os no separador «Orçamentos».
                </p>
              ) : (
                summary.categories.map((c) => {
                  const cat = catById.get(c.categoryId);
                  const budget = c.budgetCents + c.rolloverCents;
                  const pct = budget > 0 ? Math.min((c.spentCents / budget) * 100, 100) : 0;
                  const over = c.spentCents > budget;
                  return (
                    <div className="row" key={c.categoryId}>
                      <span className="dot" style={{ background: cat?.color ?? "var(--muted)" }} />
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div style={{ fontSize: 13.5, fontWeight: 600 }}>{cat?.name ?? "—"}</div>
                        <div className="track">
                          <i
                            style={{
                              width: `${pct}%`,
                              background: over ? "var(--over)" : cat?.color ?? "var(--accent)",
                            }}
                          />
                        </div>
                      </div>
                      <span
                        className={`num ${over ? "pct-over" : ""}`}
                        style={{ fontSize: 12.5, fontWeight: 700, color: "var(--muted)" }}
                      >
                        {eurosCompact(c.spentCents)} / {eurosCompact(budget)}
                      </span>
                    </div>
                  );
                })
              )}
            </div>
          </section>

          <section>
            {expected.length > 0 && (
              <>
                <p className="section-label">Previstas este mês</p>
                <div className="card">
                  {expected.map((e) => {
                    const cat = catById.get(e.categoryId);
                    return (
                      <div className="row" key={e.id}>
                        <span className="dot" style={{ background: cat?.color ?? "var(--muted)" }} />
                        <div style={{ flex: 1, minWidth: 0 }}>
                          <div style={{ fontSize: 13.5, fontWeight: 600 }}>{e.description}</div>
                          <div style={{ fontSize: 11.5, color: "var(--muted)" }}>
                            {cat?.name} · ≈ {euros(e.amountCents)}
                          </div>
                        </div>
                        <button
                          className="btn-quiet"
                          onClick={() => setSheet({ open: true, expense: e })}
                        >
                          Confirmar
                        </button>
                      </div>
                    );
                  })}
                </div>
              </>
            )}

            <p className="section-label">Recentes</p>
            <div className="card">
              {recent.length === 0 ? (
                <p style={{ color: "var(--muted)", fontSize: 13.5, padding: "12px 0" }}>
                  Sem despesas neste mês. Toca em «+ Despesa» para registar a primeira.
                </p>
              ) : (
                recent.map((e) => {
                  const cat = catById.get(e.categoryId);
                  return (
                    <button
                      key={e.id}
                      className="row"
                      style={{ width: "100%", textAlign: "left" }}
                      onClick={() => setSheet({ open: true, expense: e })}
                    >
                      <span
                        className="dot"
                        style={{ background: cat?.color ?? "var(--muted)", borderRadius: 10, width: 34, height: 34, display: "flex", alignItems: "center", justifyContent: "center", fontSize: 15, opacity: 0.9 }}
                      >
                        {cat?.icon}
                      </span>
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div style={{ fontSize: 13, fontWeight: 600 }}>
                          {e.description || cat?.name}
                        </div>
                        <div style={{ fontSize: 11, color: "var(--muted)" }}>
                          {cat?.name} ·{" "}
                          {new Date(e.spentOn + "T00:00:00").toLocaleDateString("pt-PT", {
                            day: "numeric",
                            month: "short",
                          })}
                        </div>
                      </div>
                      <span className="num" style={{ fontSize: 13, fontWeight: 700 }}>
                        −{euros(e.amountCents)}
                      </span>
                    </button>
                  );
                })
              )}
            </div>
          </section>
        </div>

        <button className="btn fab" onClick={() => setSheet({ open: true, expense: null })}>
          + Despesa
        </button>
      </main>

      {sheet.open && categories && (
        <ExpenseSheet
          categories={categories}
          initial={sheet.expense}
          onClose={() => setSheet({ open: false, expense: null })}
          onSaved={refresh}
        />
      )}
    </>
  );
}
