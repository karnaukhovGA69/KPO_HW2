CREATE TABLE IF NOT EXISTS accounts (
  id      uuid PRIMARY KEY,
  name    text NOT NULL,
  balance numeric(20,2) NOT NULL DEFAULT 0,
  CHECK (balance >= 0)
);

CREATE TABLE IF NOT EXISTS categories (
  id    uuid PRIMARY KEY,
  type  smallint NOT NULL CHECK (type IN (-1, 1)), -- -1 expense, 1 income
  name  text NOT NULL
);

CREATE TABLE IF NOT EXISTS operations (
  id              uuid PRIMARY KEY,
  type            smallint NOT NULL CHECK (type IN (-1, 1)),
  bank_account_id uuid NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
  amount          numeric(20,2) NOT NULL CHECK (amount > 0),
  "date"          date NOT NULL,
  description     text,
  category_id     uuid NOT NULL REFERENCES categories(id)
);

CREATE INDEX IF NOT EXISTS idx_ops_acc_date ON operations(bank_account_id, "date");