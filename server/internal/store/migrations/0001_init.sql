CREATE TABLE users (
    id            SERIAL PRIMARY KEY,
    username      TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE sessions (
    token      TEXT PRIMARY KEY,
    user_id    INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE categories (
    id       SERIAL PRIMARY KEY,
    name     TEXT NOT NULL,
    color    TEXT NOT NULL,
    icon     TEXT NOT NULL DEFAULT '',
    archived BOOLEAN NOT NULL DEFAULT false,
    position INT NOT NULL DEFAULT 0
);

-- Orçamentos versionados por mês de entrada em vigor: o valor em vigor num mês
-- é o da linha com effective_from mais recente <= esse mês.
CREATE TABLE budgets (
    id             SERIAL PRIMARY KEY,
    category_id    INT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    period         TEXT NOT NULL CHECK (period IN ('monthly', 'annual')),
    amount_cents   BIGINT NOT NULL CHECK (amount_cents >= 0),
    rollover       BOOLEAN NOT NULL DEFAULT false,
    effective_from DATE NOT NULL,
    UNIQUE (category_id, period, effective_from)
);

CREATE TABLE recurring_expenses (
    id                    SERIAL PRIMARY KEY,
    category_id           INT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    description           TEXT NOT NULL,
    expected_amount_cents BIGINT NOT NULL CHECK (expected_amount_cents >= 0),
    frequency             TEXT NOT NULL CHECK (frequency IN ('monthly', 'annual')),
    due_day               INT CHECK (due_day BETWEEN 1 AND 31),
    due_month             INT CHECK (due_month BETWEEN 1 AND 12),
    active                BOOLEAN NOT NULL DEFAULT true,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- status 'expected': gerada a partir de uma recorrente, ainda não confirmada;
-- só as 'confirmed' contam para o gasto real.
CREATE TABLE expenses (
    id           SERIAL PRIMARY KEY,
    category_id  INT NOT NULL REFERENCES categories(id),
    description  TEXT NOT NULL DEFAULT '',
    amount_cents BIGINT NOT NULL CHECK (amount_cents >= 0),
    spent_on     DATE NOT NULL,
    status       TEXT NOT NULL DEFAULT 'confirmed' CHECK (status IN ('expected', 'confirmed', 'skipped')),
    recurring_id INT REFERENCES recurring_expenses(id) ON DELETE SET NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX expenses_spent_on_idx ON expenses (spent_on);
CREATE INDEX expenses_category_idx ON expenses (category_id, spent_on);

-- Uma instância por recorrente por mês (impede gerar previstas em duplicado).
CREATE UNIQUE INDEX expenses_recurring_month_idx
    ON expenses (recurring_id, date_trunc('month', spent_on::timestamp))
    WHERE recurring_id IS NOT NULL;
