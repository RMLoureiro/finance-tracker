"use client";

import { useMemo, useState } from "react";
import { api, Category, Expense } from "@/lib/api";
import { currentMonth, euros } from "@/lib/money";
import { useData } from "@/lib/hooks";
import MonthNav from "@/components/MonthNav";
import ExpenseSheet from "@/components/ExpenseSheet";

export default function MovimentosPage() {
  const [month, setMonth] = useState(currentMonth());
  const [filter, setFilter] = useState<number | null>(null);
  const [sheet, setSheet] = useState<{ open: boolean; expense: Expense | null }>({
    open: false,
    expense: null,
  });

  const { data: categories } = useData(() => api.get<Category[]>("/api/categories"), []);
  const { data: expenses, reload } = useData(
    () => api.get<Expense[]>(`/api/expenses?month=${month}`),
    [month]
  );

  const catById = useMemo(() => {
    const m = new Map<number, Category>();
    (categories ?? []).forEach((c) => m.set(c.id, c));
    return m;
  }, [categories]);

  const visible = (expenses ?? []).filter(
    (e) => e.status !== "skipped" && (filter === null || e.categoryId === filter)
  );

  const byDay = useMemo(() => {
    const groups: { day: string; items: Expense[] }[] = [];
    for (const e of visible) {
      const last = groups[groups.length - 1];
      if (last && last.day === e.spentOn) last.items.push(e);
      else groups.push({ day: e.spentOn, items: [e] });
    }
    return groups;
  }, [visible]);

  const total = visible
    .filter((e) => e.status === "confirmed")
    .reduce((acc, e) => acc + e.amountCents, 0);

  return (
    <>
      <header className="head">
        <MonthNav month={month} onChange={setMonth} />
        <div style={{ marginTop: 10, fontSize: 11, textTransform: "uppercase", letterSpacing: "0.1em", opacity: 0.65 }}>
          Gasto {filter !== null ? `em ${catById.get(filter)?.name ?? ""}` : "no mês"}
        </div>
        <div className="avail num">{euros(total)}</div>
      </header>

      <main className="content">
        <div
          style={{
            display: "flex",
            gap: 8,
            overflowX: "auto",
            padding: "2px 2px 10px",
            WebkitOverflowScrolling: "touch",
          }}
        >
          <button
            className="btn-quiet"
            style={{
              whiteSpace: "nowrap",
              border: `1px solid ${filter === null ? "var(--accent)" : "var(--line)"}`,
              color: filter === null ? "var(--accent)" : "var(--muted)",
              borderRadius: 999,
              padding: "6px 14px",
              fontSize: 13,
            }}
            onClick={() => setFilter(null)}
          >
            Todas
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
                  border: `1px solid ${filter === c.id ? "var(--accent)" : "var(--line)"}`,
                  color: filter === c.id ? "var(--accent)" : "var(--muted)",
                  borderRadius: 999,
                  padding: "6px 14px",
                  fontSize: 13,
                }}
                onClick={() => setFilter(filter === c.id ? null : c.id)}
              >
                <span className="dot" style={{ background: c.color, width: 9, height: 9 }} />
                {c.name}
              </button>
            ))}
        </div>

        {byDay.length === 0 && (
          <p style={{ color: "var(--muted)", fontSize: 13.5, padding: "24px 4px" }}>
            Sem movimentos neste mês.
          </p>
        )}

        {byDay.map((g) => (
          <section key={g.day}>
            <p className="section-label" style={{ textTransform: "none", letterSpacing: 0 }}>
              {new Date(g.day + "T00:00:00").toLocaleDateString("pt-PT", {
                weekday: "long",
                day: "numeric",
                month: "long",
              })}
            </p>
            <div className="card">
              {g.items.map((e) => {
                const cat = catById.get(e.categoryId);
                const expected = e.status === "expected";
                return (
                  <button
                    key={e.id}
                    className="row"
                    style={{ width: "100%", textAlign: "left", opacity: expected ? 0.75 : 1 }}
                    onClick={() => setSheet({ open: true, expense: e })}
                  >
                    <span
                      style={{
                        background: `color-mix(in srgb, ${cat?.color ?? "#888"} 16%, transparent)`,
                        borderRadius: 10,
                        width: 34,
                        height: 34,
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "center",
                        fontSize: 15,
                        flexShrink: 0,
                      }}
                    >
                      {cat?.icon}
                    </span>
                    <div style={{ flex: 1, minWidth: 0 }}>
                      <div style={{ fontSize: 13, fontWeight: 600 }}>
                        {e.description || cat?.name}
                        {expected && (
                          <span
                            style={{
                              marginLeft: 8,
                              fontSize: 10,
                              fontWeight: 700,
                              color: "var(--accent)",
                              border: "1px solid var(--accent)",
                              borderRadius: 999,
                              padding: "1px 8px",
                            }}
                          >
                            prevista
                          </span>
                        )}
                      </div>
                      <div style={{ fontSize: 11, color: "var(--muted)" }}>{cat?.name}</div>
                    </div>
                    <span className="num" style={{ fontSize: 13, fontWeight: 700 }}>
                      {expected ? "≈ " : "−"}
                      {euros(e.amountCents)}
                    </span>
                  </button>
                );
              })}
            </div>
          </section>
        ))}

        <button className="btn fab" onClick={() => setSheet({ open: true, expense: null })}>
          + Despesa
        </button>
      </main>

      {sheet.open && categories && (
        <ExpenseSheet
          categories={categories}
          initial={sheet.expense}
          onClose={() => setSheet({ open: false, expense: null })}
          onSaved={() => {
            setSheet({ open: false, expense: null });
            reload();
          }}
        />
      )}
    </>
  );
}
