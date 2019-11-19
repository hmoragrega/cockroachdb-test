DROP DATABASE IF EXISTS bank;
CREATE DATABASE bank;
CREATE USER IF NOT EXISTS myuser;
GRANT ALL ON DATABASE bank TO myuser;
USE bank;

CREATE TABLE accounts (
    id UUID NOT NULL,
    balance_cents INT8 NOT NULL,
    CONSTRAINT "primary" PRIMARY KEY (id ASC),
    FAMILY "primary" (id, balance_cents)
);

CREATE TABLE transactions (
    account UUID NOT NULL,
    id UUID NOT NULL,
    amount_cents INT8 NOT NULL,
    description STRING NULL,
    CONSTRAINT "primary" PRIMARY KEY (account ASC, id ASC),
    CONSTRAINT fk_account FOREIGN KEY (account) REFERENCES accounts (id),
    FAMILY "primary" (account, id, amount_cents, description)
) INTERLEAVE IN PARENT accounts (account);

INSERT INTO accounts (id, balance_cents) VALUES ('90903a90-d8f0-45eb-a4aa-dea4d24b2f54', 100000); /* 1,000.00€ */
