"use client";

import { addMonths, monthLabel } from "@/lib/money";

export default function MonthNav({
  month,
  onChange,
}: {
  month: string;
  onChange: (m: string) => void;
}) {
  return (
    <div className="month-nav">
      <button onClick={() => onChange(addMonths(month, -1))} aria-label="Mês anterior">
        ‹
      </button>
      <span style={{ textTransform: "capitalize" }}>{monthLabel(month)}</span>
      <button onClick={() => onChange(addMonths(month, 1))} aria-label="Mês seguinte">
        ›
      </button>
    </div>
  );
}
