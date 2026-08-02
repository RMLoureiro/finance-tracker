"use client";

import { useState } from "react";
import { api, Category, Recurring } from "@/lib/api";
import { parseEuros } from "@/lib/money";

const MONTHS = [
  "janeiro", "fevereiro", "março", "abril", "maio", "junho",
  "julho", "agosto", "setembro", "outubro", "novembro", "dezembro",
];

interface Props {
  categories: Category[];
  initial: Recurring | null;
  onClose: () => void;
  onSaved: () => void;
}

export default function RecurringSheet({ categories, initial, onClose, onSaved }: Props) {
  const [description, setDescription] = useState(initial?.description ?? "");
  const [categoryId, setCategoryId] = useState(
    initial?.categoryId ?? categories.find((c) => !c.archived)?.id ?? 0
  );
  const [amount, setAmount] = useState(
    initial ? (initial.expectedAmountCents / 100).toFixed(2).replace(".", ",") : ""
  );
  const [frequency, setFrequency] = useState<"monthly" | "annual">(initial?.frequency ?? "monthly");
  const [dueDay, setDueDay] = useState(initial?.dueDay ? String(initial.dueDay) : "");
  const [dueMonth, setDueMonth] = useState(initial?.dueMonth ?? 1);
  const [active, setActive] = useState(initial?.active ?? true);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function save() {
    const cents = parseEuros(amount);
    if (cents === null) {
      setError("Indica o valor esperado, por exemplo 28,00");
      return;
    }
    const body = {
      categoryId,
      description,
      expectedAmountCents: cents,
      frequency,
      dueDay: dueDay === "" ? null : Number(dueDay),
      dueMonth: frequency === "annual" ? dueMonth : null,
      active,
    };
    setBusy(true);
    setError(null);
    try {
      if (initial) await api.put(`/api/recurring/${initial.id}`, body);
      else await api.post("/api/recurring", body);
      onSaved();
    } catch (e) {
      setError((e as Error).message);
      setBusy(false);
    }
  }

  async function remove() {
    if (!initial) return;
    setBusy(true);
    try {
      await api.del(`/api/recurring/${initial.id}`);
      onSaved();
    } catch (e) {
      setError((e as Error).message);
      setBusy(false);
    }
  }

  return (
    <div className="sheet-backdrop" onClick={onClose}>
      <div className="sheet" onClick={(e) => e.stopPropagation()} role="dialog" aria-modal="true">
        <h2>{initial ? "Editar despesa fixa" : "Nova despesa fixa"}</h2>
        <div className="field">
          <label htmlFor="rc-desc">Descrição</label>
          <input
            id="rc-desc"
            placeholder="Fatura da água, seguro do carro…"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
        </div>
        <div className="field">
          <label htmlFor="rc-cat">Categoria</label>
          <select id="rc-cat" value={categoryId} onChange={(e) => setCategoryId(Number(e.target.value))}>
            {categories
              .filter((c) => !c.archived || c.id === initial?.categoryId)
              .map((c) => (
                <option key={c.id} value={c.id}>
                  {c.icon} {c.name}
                </option>
              ))}
          </select>
        </div>
        <div className="field">
          <label htmlFor="rc-amount">Valor esperado (€)</label>
          <input
            id="rc-amount"
            inputMode="decimal"
            placeholder="0,00"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
          />
        </div>
        <div className="field">
          <label htmlFor="rc-freq">Frequência</label>
          <select
            id="rc-freq"
            value={frequency}
            onChange={(e) => setFrequency(e.target.value as "monthly" | "annual")}
          >
            <option value="monthly">Mensal</option>
            <option value="annual">Anual</option>
          </select>
        </div>
        {frequency === "annual" && (
          <div className="field">
            <label htmlFor="rc-month">Mês da despesa</label>
            <select id="rc-month" value={dueMonth} onChange={(e) => setDueMonth(Number(e.target.value))}>
              {MONTHS.map((m, i) => (
                <option key={m} value={i + 1}>
                  {m}
                </option>
              ))}
            </select>
          </div>
        )}
        <div className="field">
          <label htmlFor="rc-day">Dia esperado (opcional)</label>
          <input
            id="rc-day"
            inputMode="numeric"
            placeholder="por exemplo, 12"
            value={dueDay}
            onChange={(e) => setDueDay(e.target.value.replace(/\D/g, ""))}
          />
        </div>
        {initial && (
          <div className="field">
            <label style={{ display: "flex", alignItems: "center", gap: 8, color: "var(--ink)" }}>
              <input
                type="checkbox"
                style={{ width: "auto" }}
                checked={active}
                onChange={(e) => setActive(e.target.checked)}
              />
              Ativa (gera despesas previstas)
            </label>
          </div>
        )}
        {error && <p className="error">{error}</p>}
        <div style={{ display: "flex", gap: 8, marginTop: 14 }}>
          <button className="btn" style={{ flex: 1 }} onClick={save} disabled={busy}>
            Guardar
          </button>
          {initial && (
            <button className="btn-danger" onClick={remove} disabled={busy}>
              Apagar
            </button>
          )}
          <button className="btn-quiet" onClick={onClose} disabled={busy}>
            Cancelar
          </button>
        </div>
      </div>
    </div>
  );
}
