"use client";

import { useEffect, useRef, useState } from "react";
import { euros } from "@/lib/money";

export interface BarDatum {
  label: string;
  value: number;
  /** referência (orçamento) para este ponto, opcional */
  target?: number;
}

const W = 640;
const H = 220;
const PAD_TOP = 26;
const PAD_BOTTOM = 24;

/**
 * Gráfico de barras de série única (valores em cêntimos), com linha de
 * referência opcional para o orçamento. Barras finas com topo arredondado,
 * rótulos diretos apenas no ponto em foco.
 */
export default function BarChart({ data, color }: { data: BarDatum[]; color: string }) {
  const [focus, setFocus] = useState<number | null>(null);
  const scroller = useRef<HTMLDivElement>(null);

  // em ecrãs estreitos o gráfico faz scroll horizontal; começa no fim,
  // onde estão os meses mais recentes
  useEffect(() => {
    const el = scroller.current;
    if (el) el.scrollLeft = el.scrollWidth;
  }, []);

  const max = Math.max(...data.map((d) => Math.max(d.value, d.target ?? 0)), 1);
  const innerH = H - PAD_TOP - PAD_BOTTOM;
  const step = W / data.length;
  const barW = Math.min(step * 0.55, 34);

  const y = (v: number) => PAD_TOP + innerH * (1 - v / max);

  return (
    <div ref={scroller} style={{ overflowX: "auto" }}>
      <svg
        viewBox={`0 0 ${W} ${H}`}
        style={{ width: "100%", minWidth: 480, display: "block" }}
        role="img"
        aria-label="Gasto por mês"
        onMouseLeave={() => setFocus(null)}
      >
        {/* linha de base */}
        <line x1="0" y1={H - PAD_BOTTOM} x2={W} y2={H - PAD_BOTTOM} stroke="var(--line)" />

        {data.map((d, i) => {
          const x = i * step + (step - barW) / 2;
          const h = Math.max(innerH * (d.value / max), d.value > 0 ? 2 : 0);
          const focused = focus === i;
          return (
            <g key={d.label}>
              {/* alvo de rato maior do que a barra */}
              <rect
                x={i * step}
                y={0}
                width={step}
                height={H}
                fill="transparent"
                onMouseEnter={() => setFocus(i)}
              />
              <rect
                x={x}
                y={y(d.value)}
                width={barW}
                height={h}
                rx={4}
                fill={color}
                opacity={focus === null || focused ? 1 : 0.45}
                style={{ pointerEvents: "none" }}
              />
              {d.target !== undefined && d.target > 0 && (
                <line
                  x1={i * step + step * 0.12}
                  x2={(i + 1) * step - step * 0.12}
                  y1={y(d.target)}
                  y2={y(d.target)}
                  stroke="var(--muted)"
                  strokeWidth="1.5"
                  strokeDasharray="4 3"
                  style={{ pointerEvents: "none" }}
                />
              )}
              <text
                x={i * step + step / 2}
                y={H - 7}
                textAnchor="middle"
                fontSize="11"
                fill={focused ? "var(--ink)" : "var(--muted)"}
                style={{ pointerEvents: "none" }}
              >
                {d.label}
              </text>
              {focused && (
                <text
                  x={Math.min(Math.max(i * step + step / 2, 40), W - 40)}
                  y={Math.max(y(d.value) - 8, 14)}
                  textAnchor="middle"
                  fontSize="12"
                  fontWeight="700"
                  fill="var(--ink)"
                  style={{ pointerEvents: "none" }}
                >
                  {euros(d.value)}
                </text>
              )}
            </g>
          );
        })}
      </svg>
      <p style={{ fontSize: 11.5, color: "var(--muted)", margin: "4px 2px 0" }}>
        Barras: gasto do mês · linha tracejada: orçamento em vigor
      </p>
    </div>
  );
}
