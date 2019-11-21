package transactions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Errors
var (
	ErrBeginTx           = errors.New("could not start transaction")
	ErrUpdateBalance     = errors.New("could not update balance")
	ErrInsertTx          = errors.New("could not insert transaction")
	ErrScanBalance       = errors.New("could not scan the balance")
	ErrScanAccount       = errors.New("could not scan an account")
	ErrCreateAccount     = errors.New("could not create an account")
	ErrDeleteAccount     = errors.New("could not delete an account")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrCommitTx          = errors.New("could not commit a transaction")
	ErrRollbackTx        = errors.New("could not rollback a transaction")
	ErrRowsAffected      = errors.New("could not detect the affected rows")
	ErrDeleteNotFound    = errors.New("could not found the account to delete")
)

type Transaction struct {
	AmountCents int64
	Description string
}

type Executor interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type TxExecutor interface {
	Executor
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type Accounting struct {
	executor TxExecutor
}

func NewAccounting(executor TxExecutor) *Accounting {
	return &Accounting{executor}
}

func readBalance(ctx context.Context, e Executor, accountID string) (int64, error) {
	row := e.QueryRowContext(ctx, "SELECT balance_cents FROM accounts WHERE id = $1", accountID)

	var balance int64
	if err := row.Scan(&balance); err != nil {
		return 0, fmt.Errorf("%w %s: %v", ErrScanBalance, accountID, err)
	}

	return balance, nil
}

func (a *Accounting) ReadBalance(ctx context.Context, accountID string) (int64, error) {
	 return readBalance(ctx, a.executor, accountID)
}

func (a *Accounting) CreateAccount(ctx context.Context, balance int64) (string, error) {
 	row := a.executor.QueryRowContext(ctx, "INSERT INTO accounts (balance_cents) VALUES ($1) RETURNING id", balance)

	var accountID string
 	if err := row.Scan(&accountID); err != nil {
		return "", fmt.Errorf("%w: %v", ErrCreateAccount, err)
	}

 	return accountID, nil
}

func (a *Accounting) DeleteAccount(ctx context.Context, accountID string) error {
 	r, err := a.executor.ExecContext(ctx, "DELETE FROM accounts WHERE id = $1", accountID)
 	if err != nil {
		return fmt.Errorf("%w %s: %v", ErrDeleteAccount, accountID, err)
	}

	c, err := r.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w %s: %v", ErrRowsAffected, accountID, err)
	}

	if c != 1 {
		return fmt.Errorf("%w %s: %v", ErrDeleteNotFound, accountID, err)
	}

 	return nil
}

func (a *Accounting) RandomAccount(ctx context.Context) (string, int64, error) {
	row := a.executor.QueryRowContext(ctx, "SELECT id, balance_cents FROM accounts ORDER BY random()")

	var accountID string
	var balance int64
	if err := row.Scan(&accountID, &balance); err != nil {
		return "", 0, fmt.Errorf("%w: %v", ErrScanAccount, err)
	}

	return accountID, balance, nil
}

func (a *Accounting) ProcessList(ctx context.Context, accountID string, txs []Transaction, parallelize bool) (int64, error) {
	dbTx, err := a.executor.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrBeginTx, err)
	}

	extraSQL := ""
	if parallelize {
		extraSQL += " RETURNING NOTHING"
	}

	insertSQL := `INSERT INTO transactions (account, id, amount_cents, description) 
			VALUES ($1, gen_random_uuid(), $2, $3)` + extraSQL

	var total int64
	for _, tx := range txs {
		if _, err := dbTx.ExecContext(ctx, insertSQL, accountID, tx.AmountCents, tx.Description); err != nil {
			return 0, rollback(dbTx, fmt.Errorf("%w: %v", ErrInsertTx, err))
		}

		total += tx.AmountCents
	}

	_, err = dbTx.Exec(
		`UPDATE accounts SET balance_cents = balance_cents - $1 WHERE id = $2`+extraSQL,
		total, accountID,
	)
	if err != nil {
		return 0, rollback(dbTx, fmt.Errorf("%w: %v", ErrUpdateBalance, err))
	}

	balance, err := readBalance(ctx, dbTx, accountID)
	if err != nil {
		return 0, rollback(dbTx, err)
	}

	if balance < 0 {
		return 0, rollback(dbTx, fmt.Errorf("%w (funds %.2f on account ID %s): %v", ErrInsufficientFunds, float64(balance/100), accountID, err))
	}

	if err := dbTx.Commit(); err != nil {
		return 0, fmt.Errorf("%w on %s: %v", ErrCommitTx, accountID, err)
	}

	return balance, nil
}

func rollback(dbTx *sql.Tx, wrappingError error) error {
	if err := dbTx.Rollback(); err != nil {
		return fmt.Errorf("%v: %w", ErrRollbackTx, wrappingError)
	} else {
		return wrappingError
	}
}