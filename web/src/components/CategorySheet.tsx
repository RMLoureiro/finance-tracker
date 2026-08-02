"use client";

import { useState } from "react";
import { api, Category } from "@/lib/api";

// Paleta validada para daltonismo (ver README do projeto).
const PALETTE = ["#4A8AC9", "#B07E1F", "#0F9488", "#C25E37", "#8A64B8", "#C94F72"];

interface Props {
  initial: Category | null;
  onClose: () => void;
  onSaved: () => void;
}

export default function CategorySheet({ initial, onClose, onSaved }: Props) {
  const [name, setName] = useState(initial?.name ?? "");
  const [color, setColor] = useState(initial?.color ?? PALETTE[0]);
  const [icon, setIcon] = useState(initial?.icon ?? "");
  const [archived, setArchived] = useState(initial?.archived ?? false);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function save() {
    setBusy(true);
    setError(null);
    const body = { name, color, icon, archived };
    try {
      if (initial) await api.put(`/api/categories/${initial.id}`, body);
      else await api.post("/api/categories", body);
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
      await api.del(`/api/categories/${initial.id}`);
      onSaved();
    } catch (e) {
      setError((e as Error).message);
      setBusy(false);
    }
  }

  return (
    <div className="sheet-backdrop" onClick={onClose}>
      <div className="sheet" onClick={(e) => e.stopPropagation()} role="dialog" aria-modal="true">
        <h2>{initial ? "Editar categoria" : "Nova categoria"}</h2>
        <div className="field">
          <label htmlFor="ct-name">Nome</label>
          <input id="ct-name" value={name} onChange={(e) => setName(e.target.value)} />
        </div>
        <div className="field">
          <label>Cor</label>
          <div style={{ display: "flex", gap: 10 }}>
            {PALETTE.map((c) => (
              <button
                key={c}
                aria-label={`Cor ${c}`}
                onClick={() => setColor(c)}
                style={{
                  width: 34,
                  height: 34,
                  borderRadius: 10,
                  background: c,
                  border: color === c ? "3px solid var(--ink)" : "3px solid transparent",
                }}
              />
            ))}
          </div>
        </div>
        <div className="field">
          <label htmlFor="ct-icon">Ícone (emoji)</label>
          <input
            id="ct-icon"
            value={icon}
            onChange={(e) => setIcon(e.target.value)}
            placeholder="💧 💡 ⛽ 🐾 …"
            maxLength={4}
          />
        </div>
        {initial && (
          <div className="field">
            <label style={{ display: "flex", alignItems: "center", gap: 8, color: "var(--ink)" }}>
              <input
                type="checkbox"
                style={{ width: "auto" }}
                checked={archived}
                onChange={(e) => setArchived(e.target.checked)}
              />
              Arquivada (deixa de aparecer em novas despesas)
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
