"use client";

import { useMemo, useState } from "react";
import { api, AnnualCategoryReport, Category, ReportMonth } from "@/lib/api";
import { currentMonth, euros, eurosCompact } from "@/lib/money";
import { useData } from "@/lib/hooks";
import BarChart from "@/components/BarChart";

const MONTH_SHORT = ["jan", "fev", "mar", "abr", "mai", "jun", "jul", "ago", "set", "out", "nov", "dez"];

function shortLabel(month: string): string {
  const [y, m] = month.split("-").map(Number);
  return m === 1 ? `${MONTH_SHORT[0]} ${String(y).slice(2)}` : MONTH_SHORT[m - 1];
}

export default function RelatoriosPage() {
  const [categoryId, setCategoryId] = useState<number | null>(null);
  const [year, setYear] = useState(new Date().getFullYear());

  const { data: categories } = useData(() => api.get<Category[]>("/api/categories"), []);
  const { data: monthly } = useData(
    () => api.get<ReportMonth[]>(`/api/reports/monthly?months=13&end=${currentMonth()}`),
    []
  );
  const { data: annual } = useData(
    () => api.get<AnnualCategoryReport[]>(`/api/reports/annual?year=${year}`),
    [year]
  );

  const catById = useMemo(() => {
    const m = new Map<number, Category>();
    (categories ?? []).forEach((c) => m.set(c.id, c));
    return m;
  }, [categories]);

  const selected = categoryId !== null ? catById.get(categoryId) : null;

  const chartData = useMemo(() => {
    const rows = monthly ?? [];
    // 13 meses pedidos: os últimos 12 vão ao gráfico; o primeiro serve para
    // a comparação homóloga
    return rows.slice(-12).map((r) => ({
      label: shortLabel(r.month),
      value: categoryId === null ? r.spentCents : (r.byCategory[String(categoryId)] ?? 0),
      target: categoryId === null ? r.budgetCents : undefined,
    }));
  }, [monthly, categoryId]);

  const comparison = useMemo(() => {
    const rows = monthly ?? [];
    if (rows.length < 13) return null;
    const value = (r: ReportMonth) =>
      categoryId === null ? r.spentCents : (r.byCategory[String(categoryId)] ?? 0);
    const nowRow = rows[rows.length - 1];
    const lastYear = rows[0];
    const past = rows.slice(0, -1);
    const avg = past.reduce((acc, r) => acc + value(r), 0) / past.length;
    return { now: value(nowRow), lastYear: value(lastYear), avg: Math.round(avg) };
  }, [monthly, categoryId]);

  const annualTotal = (annual ?? []).reduce(
    (acc, r) => ({ spent: acc.spent + r.spentCents, budget: acc.budget + r.budgetCents }),
    { spent: 0, budget: 0 }
  );

  return (
    <main className="content" style={{ paddingTop: 20 }}>
      <p className="section-label">Evolução mensal</p>

      <div style={{ display: "flex", gap: 8, overflowX: "auto", padding: "2px 2px 10px" }}>
        <button
          className="btn-quiet"
          style={{
            whiteSpace: "nowrap",
            border: `1px solid ${categoryId === null ? "var(--accent)" : "var(--line)"}`,
            color: categoryId === null ? "var(--accent)" : "var(--muted)",
            borderRadius: 999,
            padding: "6px 14px",
            fontSize: 13,
          }}
          onClick={() => setCategoryId(null)}
        >
          Total
        </button>
        {(categories ?? [])
          .filter((c) => !c.archived)
          .map((c) => (
            <button
              key={c.id}
              className="btn-quiet"
              style={{
                whiteSpace: "nowrap",
                display: "inline-flex",
                alignItems: "center",
                gap: 6,
                border: `1px solid ${categoryId === c.id ? "var(--accent)" : "var(--line)"}`,
                color: categoryId === c.id ? "var(--accent)" : "var(--muted)",
                borderRadius: 999,
                padding: "6px 14px",
                fontSize: 13,
              }}
              onClick={() => setCategoryId(categoryId === c.id ? null : c.id)}
            >
              <span className="dot" style={{ background: c.color, width: 9, height: 9 }} />
              {c.name}
            </button>
          ))}
      </div>

      <div className="card" style={{ padding: "14px 14px 10px" }}>
        {chartData.length > 0 ? (
          <BarChart data={chartData} color={selected?.color ?? "var(--accent)"} />
        ) : (
          <p style={{ color: "var(--muted)", fontSize: 13.5, padding: "12px 0" }}>Sem dados.</p>
        )}
      </div>

      {comparison && (
        <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 12 }}>
          {[
            { label: "Este mês", value: comparison.now },
            { label: "Média 12 meses", value: comparison.avg },
            { label: "Há um ano", value: comparison.lastYear },
          ].map((t) => (
            <div key={t.label} className="card" style={{ padding: "12px 14px", margin: 0 }}>
              <div style={{ fontSize: 11, textTransform: "uppercase", letterSpacing: "0.08em", color: "var(--muted)", fontWeight: 700 }}>
                {t.label}
              </div>
              <div className="num" style={{ fontSize: 18, fontWeight: 700, marginTop: 2 }}>
                {euros(t.value)}
              </div>
            </div>
          ))}
        </div>
      )}

      <div style={{ display: "flex", alignItems: "baseline", justifyContent: "space-between", marginTop: 20 }}>
        <p className="section-label" style={{ margin: "0 4px 8px" }}>
          Resumo anual
        </p>
        <div style={{ display: "flex", gap: 4, alignItems: "center", fontSize: 13.5 }}>
          <button className="btn-quiet" style={{ padding: "4px 10px" }} onClick={() => setYear(year - 1)} aria-label="Ano anterior">
            ‹
          </button>
          <span className="num" style={{ fontWeight: 700 }}>{year}</span>
          <button className="btn-quiet" style={{ padding: "4px 10px" }} onClick={() => setYear(year + 1)} aria-label="Ano seguinte">
            ›
          </button>
        </div>
      </div>

      <div className="card">
        {(annual ?? []).length === 0 && (
          <p style={{ color: "var(--muted)", fontSize: 13.5, padding: "12px 0" }}>
            Sem despesas em {year}.
          </p>
        )}
        {(annual ?? []).map((r) => {
          const cat = catById.get(r.categoryId);
          const pct = r.budgetCents > 0 ? Math.min((r.spentCents / r.budgetCents) * 100, 100) : 0;
          const over = r.budgetCents > 0 && r.spentCents > r.budgetCents;
          return (
            <div className="row" key={r.categoryId}>
              <span className="dot" style={{ background: cat?.color ?? "var(--muted)" }} />
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontSize: 13.5, fontWeight: 600 }}>{cat?.name ?? "—"}</div>
                {r.budgetCents > 0 && (
                  <div className="track">
                    <i style={{ width: `${pct}%`, background: over ? "var(--over)" : cat?.color ?? "var(--accent)" }} />
                  </div>
                )}
              </div>
              <span className={`num ${over ? "pct-over" : ""}`} style={{ fontSize: 12.5, fontWeight: 700, color: "var(--muted)" }}>
                {eurosCompact(r.spentCents)}
                {r.budgetCents > 0 && ` / ${eurosCompact(r.budgetCents)}`}
              </span>
            </div>
          );
        })}
        {(annual ?? []).length > 0 && (
          <div className="row" style={{ fontWeight: 700 }}>
            <span className="dot" style={{ background: "transparent" }} />
            <div style={{ flex: 1, fontSize: 13.5 }}>Total</div>
            <span className="num" style={{ fontSize: 12.5 }}>
              {eurosCompact(annualTotal.spent)}
              {annualTotal.budget > 0 && ` / ${eurosCompact(annualTotal.budget)}`}
            </span>
          </div>
        )}
      </div>
    </main>
  );
}
