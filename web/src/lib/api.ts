export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

export interface Category {
  id: number;
  name: string;
  color: string;
  icon: string;
  archived: boolean;
  position: number;
}

export interface Budget {
  categoryId: number;
  period: "monthly" | "annual";
  amountCents: number;
  rollover: boolean;
  effectiveFrom: string;
}

export interface Expense {
  id: number;
  categoryId: number;
  description: string;
  amountCents: number;
  spentOn: string;
  status: "expected" | "confirmed" | "skipped";
  recurringId: number | null;
}

export interface Recurring {
  id: number;
  categoryId: number;
  description: string;
  expectedAmountCents: number;
  frequency: "monthly" | "annual";
  dueDay: number | null;
  dueMonth: number | null;
  active: boolean;
}

export interface MonthCategory {
  categoryId: number;
  monthlyBudgetCents: number;
  annualBudgetCents: number;
  dilutedCents: number;
  budgetCents: number;
  rollover: boolean;
  rolloverCents: number;
  spentCents: number;
  expectedCents: number;
  availableCents: number;
}

export interface MonthSummary {
  month: string;
  daysInMonth: number;
  dayOfMonth: number | null;
  projectedCents: number | null;
  totals: {
    budgetCents: number;
    rolloverCents: number;
    spentCents: number;
    expectedCents: number;
    availableCents: number;
  };
  categories: MonthCategory[];
}

export interface ReportMonth {
  month: string;
  spentCents: number;
  budgetCents: number;
  byCategory: Record<string, number>;
}

export interface AnnualCategoryReport {
  categoryId: number;
  spentCents: number;
  budgetCents: number;
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
  const resp = await fetch(path, {
    method,
    credentials: "same-origin",
    headers: body !== undefined ? { "Content-Type": "application/json" } : undefined,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  if (resp.status === 401 && !location.pathname.startsWith("/login")) {
    location.href = "/login";
    throw new ApiError(401, "sessão expirada");
  }

  let data: unknown = null;
  try {
    data = await resp.json();
  } catch {
    /* respostas sem corpo */
  }

  if (!resp.ok) {
    const msg =
      data && typeof data === "object" && "error" in data
        ? String((data as { error: string }).error)
        : `erro ${resp.status}`;
    throw new ApiError(resp.status, msg);
  }
  return data as T;
}

export const api = {
  get: <T>(path: string) => request<T>("GET", path),
  post: <T>(path: string, body: unknown) => request<T>("POST", path, body),
  put: <T>(path: string, body: unknown) => request<T>("PUT", path, body),
  del: <T>(path: string) => request<T>("DELETE", path),
};
