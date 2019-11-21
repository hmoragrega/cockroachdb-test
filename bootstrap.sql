DROP DATABASE IF EXISTS bank;
CREATE DATABASE bank;
CREATE USER IF NOT EXISTS myuser;
GRANT ALL ON DATABASE bank TO myuser;
USE bank;

CREATE TABLE accounts (
    id UUID NOT NULL DEFAULT gen_random_uuid(),
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
    CONSTRAINT fk_account FOREIGN KEY (account) REFERENCES accounts (id) ON DELETE CASCADE,
    FAMILY "primary" (account, id, amount_cents, description)
) INTERLEAVE IN PARENT accounts (account);

INSERT INTO accounts (id, balance_cents) VALUES
('90903a90-d8f0-45eb-a4aa-dea4d24b2f54', 100000), /* 1,000.00€ */
('704855f3-b9cf-4496-8168-51b61181f323', 100000), /* 1,000.00€ */
('e6b9caaa-b4f5-44c4-81d1-d737c9e2804d', 100000), /* 1,000.00€ */
('aa49a2b1-0bc0-42c0-8afa-82257fac3eb5', 100000), /* 1,000.00€ */
('742ff78b-c7ff-4796-ba69-5e2c46329120', 100000), /* 1,000.00€ */
('881e0b7b-91dd-4e62-9960-4796ece00379', 100000), /* 1,000.00€ */
('8d927bd7-120e-4fc8-ad31-00a112035835', 100000), /* 1,000.00€ */
('e2c343b5-44ff-4b98-b0c3-311ea84d460e', 100000), /* 1,000.00€ */
('7224e162-6862-4317-9ead-1ad8b9d4ce46', 100000), /* 1,000.00€ */
('661dcb51-c1f0-465e-8dab-5abdfd8a7c23', 100000); /* 1,000.00€ */
