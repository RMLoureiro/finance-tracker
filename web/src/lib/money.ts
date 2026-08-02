const fmt = new Intl.NumberFormat("pt-PT", {
  style: "currency",
  currency: "EUR",
});

const fmtWhole = new Intl.NumberFormat("pt-PT", {
  style: "currency",
  currency: "EUR",
  maximumFractionDigits: 0,
});

export function euros(cents: number): string {
  return fmt.format(cents / 100);
}

/** Formato compacto para listas: sem cêntimos quando são zero. */
export function eurosCompact(cents: number): string {
  return cents % 100 === 0 ? fmtWhole.format(cents / 100) : fmt.format(cents / 100);
}

/** Converte "12,34" / "12.34" / "12" em cêntimos; null se inválido. */
export function parseEuros(input: string): number | null {
  const clean = input.trim().replace(/\s|€/g, "").replace(",", ".");
  if (!/^\d+(\.\d{1,2})?$/.test(clean)) return null;
  return Math.round(parseFloat(clean) * 100);
}

/** "2026-08" → "Agosto de 2026" */
export function monthLabel(month: string): string {
  const [y, m] = month.split("-").map(Number);
  const label = new Date(Date.UTC(y, m - 1, 1)).toLocaleDateString("pt-PT", {
    month: "long",
    year: "numeric",
    timeZone: "UTC",
  });
  return label.charAt(0).toUpperCase() + label.slice(1);
}

export function currentMonth(): string {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}`;
}

export function addMonths(month: string, delta: number): string {
  const [y, m] = month.split("-").map(Number);
  const d = new Date(Date.UTC(y, m - 1 + delta, 1));
  return `${d.getUTCFullYear()}-${String(d.getUTCMonth() + 1).padStart(2, "0")}`;
}

export function todayISO(): string {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-${String(now.getDate()).padStart(2, "0")}`;
}
