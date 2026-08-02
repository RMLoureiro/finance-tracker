"use client";

import { useState } from "react";
import { api, Budget, Category } from "@/lib/api";
import { currentMonth, parseEuros } from "@/lib/money";

interface Props {
  category: Category;
  monthly: Budget | null;
  annual: Budget | null;
  onClose: () => void;
  onSaved: () => void;
}

function toInput(cents: number | undefined): string {
  if (!cents) return "";
  return (cents / 100).toFixed(2).replace(".", ",");
}

export default function BudgetSheet({ category, monthly, annual, onClose, onSaved }: Props) {
  const [monthlyAmount, setMonthlyAmount] = useState(toInput(monthly?.amountCents));
  const [annualAmount, setAnnualAmount] = useState(toInput(annual?.amountCents));
  const [rollover, setRollover] = useState(monthly?.rollover ?? false);
  const [from, setFrom] = useState(currentMonth());
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function save() {
    const m = monthlyAmount.trim() === "" ? 0 : parseEuros(monthlyAmount);
    const a = annualAmount.trim() === "" ? 0 : parseEuros(annualAmount);
    if (m === null || a === null) {
      setError("Valores inválidos — usa por exemplo 65 ou 65,00");
      return;
    }
    setBusy(true);
    setError(null);
    try {
      await api.put("/api/budgets", {
        categoryId: category.id,
        period: "monthly",
        amountCents: m,
        rollover,
        month: from,
      });
      await api.put("/api/budgets", {
        categoryId: category.id,
        period: "annual",
        amountCents: a,
        rollover: false,
        month: from,
      });
      onSaved();
    } catch (e) {
      setError((e as Error).message);
      setBusy(false);
    }
  }

  return (
    <div className="sheet-backdrop" onClick={onClose}>
      <div className="sheet" onClick={(e) => e.stopPropagation()} role="dialog" aria-modal="true">
        <h2>
          {category.icon} Orçamento — {category.name}
        </h2>
        <div className="field">
          <label htmlFor="bg-monthly">Mensal (€)</label>
          <input
            id="bg-monthly"
            inputMode="decimal"
            placeholder="0,00"
            value={monthlyAmount}
            onChange={(e) => setMonthlyAmount(e.target.value)}
          />
        </div>
        <div className="field">
          <label htmlFor="bg-annual">Anual (€) — aparece diluído a 1/12 em cada mês</label>
          <input
            id="bg-annual"
            inputMode="decimal"
            placeholder="0,00"
            value={annualAmount}
            onChange={(e) => setAnnualAmount(e.target.value)}
          />
        </div>
        <div className="field">
          <label style={{ display: "flex", alignItems: "center", gap: 8, color: "var(--ink)" }}>
            <input
              type="checkbox"
              style={{ width: "auto" }}
              checked={rollover}
              onChange={(e) => setRollover(e.target.checked)}
            />
            Transitar saldo entre meses (rollover)
          </label>
        </div>
        <div className="field">
          <label htmlFor="bg-from">Aplicar a partir de</label>
          <input id="bg-from" type="month" value={from} onChange={(e) => setFrom(e.target.value)} />
        </div>
        {error && <p className="error">{error}</p>}
        <div style={{ display: "flex", gap: 8, marginTop: 14 }}>
          <button className="btn" style={{ flex: 1 }} onClick={save} disabled={busy}>
            Guardar
          </button>
          <button className="btn-quiet" onClick={onClose} disabled={busy}>
            Cancelar
          </button>
        </div>
      </div>
    </div>
  );
}
