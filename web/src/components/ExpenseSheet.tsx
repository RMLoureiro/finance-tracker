"use client";

import { useState } from "react";
import { api, Category, Expense } from "@/lib/api";
import { euros, parseEuros, todayISO } from "@/lib/money";

interface Props {
  categories: Category[];
  /** despesa a editar/confirmar; null para criar nova */
  initial: Expense | null;
  onClose: () => void;
  onSaved: () => void;
}

export default function ExpenseSheet({ categories, initial, onClose, onSaved }: Props) {
  const confirming = initial?.status === "expected";
  const [amount, setAmount] = useState(
    initial ? (initial.amountCents / 100).toFixed(2).replace(".", ",") : ""
  );
  const [categoryId, setCategoryId] = useState<number>(
    initial?.categoryId ?? categories.find((c) => !c.archived)?.id ?? 0
  );
  const [description, setDescription] = useState(initial?.description ?? "");
  const [spentOn, setSpentOn] = useState(initial?.spentOn ?? todayISO());
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const active = categories.filter((c) => !c.archived || c.id === initial?.categoryId);

  async function save() {
    const cents = parseEuros(amount);
    if (cents === null || cents === 0) {
      setError("Indica um valor válido, por exemplo 12,50");
      return;
    }
    setBusy(true);
    setError(null);
    const body = {
      categoryId,
      description,
      amountCents: cents,
      spentOn,
      status: "confirmed",
    };
    try {
      if (initial) {
        await api.put(`/api/expenses/${initial.id}`, body);
      } else {
        await api.post("/api/expenses", body);
      }
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
      await api.del(`/api/expenses/${initial.id}`);
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
          {confirming
            ? "Confirmar despesa prevista"
            : initial
              ? "Editar despesa"
              : "Nova despesa"}
        </h2>
        {confirming && initial && (
          <p style={{ marginTop: -8, color: "var(--muted)", fontSize: 13 }}>
            Valor previsto: {euros(initial.amountCents)} — ajusta para o valor real da fatura.
          </p>
        )}
        <div className="field">
          <label htmlFor="exp-amount">Valor (€)</label>
          <input
            id="exp-amount"
            inputMode="decimal"
            placeholder="0,00"
            autoFocus={!initial}
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
          />
        </div>
        <div className="field">
          <label htmlFor="exp-cat">Categoria</label>
          <select id="exp-cat" value={categoryId} onChange={(e) => setCategoryId(Number(e.target.value))}>
            {active.map((c) => (
              <option key={c.id} value={c.id}>
                {c.icon} {c.name}
              </option>
            ))}
          </select>
        </div>
        <div className="field">
          <label htmlFor="exp-desc">Descrição (opcional)</label>
          <input
            id="exp-desc"
            placeholder="Continente, fatura da luz…"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
        </div>
        <div className="field">
          <label htmlFor="exp-date">Data</label>
          <input id="exp-date" type="date" value={spentOn} onChange={(e) => setSpentOn(e.target.value)} />
        </div>
        {error && <p className="error">{error}</p>}
        <div style={{ display: "flex", gap: 8, marginTop: 14 }}>
          <button className="btn" style={{ flex: 1 }} onClick={save} disabled={busy}>
            {confirming ? "Confirmar" : "Guardar"}
          </button>
          {initial && !confirming && (
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
