package transactions

import (
	"database/sql"
	"fmt"
)

type Transaction struct {
	AmountCents int64
	Description string
}

func ProcessList(db *sql.DB, accountID int64, txs []Transaction, parallelize bool) error {
	dbTx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not start transaction: %v", err)
	}

	extraSQL := ""
	if parallelize {
		extraSQL += " RETURNING NOTHING"
	}

	insertSQL := `INSERT INTO transactions (account, id, amount_cents, description) 
			VALUES ($1, gen_random_uuid(), $2, $3)` + extraSQL

	var total int64
	for _, tx := range txs {
		if _, err := dbTx.Exec(insertSQL, accountID, tx.AmountCents, tx.Description); err != nil {
			return rollback(dbTx, fmt.Errorf("could not insert transaction: %v", err))
		}

		total += tx.AmountCents
	}

	_, err = dbTx.Exec(
		`UPDATE accounts SET balance_cents = balance_cents - $1 WHERE id = $2`+extraSQL,
		total, accountID,
	)
	if err != nil {
		return rollback(dbTx, fmt.Errorf("could not update balance: %v", err))
	}

	var balance int64
	row := dbTx.QueryRow("SELECT balance_cents FROM accounts WHERE id = $1", accountID)
	if err := row.Scan(&balance); err != nil {
		return rollback(dbTx, fmt.Errorf("could not scan balance: %v", err))
	}

	if balance < 0 {
		return rollback(dbTx, fmt.Errorf("insufficient funds %.2f on account ID %d", float64(balance/100), accountID))
	}

	if err := dbTx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %v", err)
	}

	return nil
}

func rollback(dbTx *sql.Tx, wrappingError error) error {
	if err := dbTx.Rollback(); err != nil {
		return fmt.Errorf("could not rollback (wrapped error: %v): %v", wrappingError, err)
	} else {
		return wrappingError
	}
}
