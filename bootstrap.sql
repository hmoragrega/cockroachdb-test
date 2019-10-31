DROP DATABASE IF EXISTS bank;
CREATE DATABASE bank;
CREATE USER IF NOT EXISTS myuser;
GRANT ALL ON DATABASE bank TO myuser;
USE bank;

CREATE TABLE accounts (
    id INT8 NOT NULL,
    balance_cents INT8 NOT NULL,
    CONSTRAINT "primary" PRIMARY KEY (id ASC),
    FAMILY "primary" (id, balance_cents)
);

CREATE TABLE transactions (
    account INT8 NOT NULL,
    id UUID NOT NULL,
    amount_cents INT8 NOT NULL,
    description STRING NULL,
    CONSTRAINT "primary" PRIMARY KEY (account ASC, id ASC),
    CONSTRAINT fk_account FOREIGN KEY (account) REFERENCES accounts (id),
    FAMILY "primary" (account, id, amount_cents, description)
) INTERLEAVE IN PARENT accounts (account);

INSERT INTO accounts (id, balance_cents) VALUES (1, 100000); /* 1,000.00€ */
