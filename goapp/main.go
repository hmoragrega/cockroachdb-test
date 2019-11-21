package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"goapp/transactions"
	"github.com/smira/go-statsd"
	"github.com/google/uuid"
	"math/rand"

	// "goapp/transactions"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var appId string
var statsdC *statsd.Client
var r *rand.Rand

func reads(ctx context.Context, a *transactions.Accounting, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	out:
	for {
		select {
		case <-ctx.Done():
			logInfo("quitting routine")
			break out
		default:
			accountID, _, err := a.RandomAccount(ctx)
			monitor("get_random_account", err)
			if err != nil {
				break
			}

			_, err = a.ReadBalance(ctx, accountID)
			monitor("read_balance", err, accountID)
		}
	}
}

func txs(ctx context.Context, a *transactions.Accounting, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	out:
	for {
		select {
		case <-ctx.Done():
			logInfo("quitting routine")
			break out
		default:
			accountID, balance, err := a.RandomAccount(ctx)
			monitor("get_random_account", err)
			if err != nil {
				break
			}

			list := generateRandomTransactions(balance)

			balance, err = a.ProcessList(ctx, accountID, list, true)
			monitor("process_transactions", err, fmt.Sprintf("current balance: %d", balance))
			if err != nil {
				break
			}
		}
	}
}

func deletes(ctx context.Context, a *transactions.Accounting, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

out:
	for {
		select {
		case <-ctx.Done():
			logInfo("quitting routine")
			break out
		default:
			accountID, _, err := a.RandomAccount(ctx)
			monitor("get_random_account", err)
			if err != nil {
				break
			}

			err = a.DeleteAccount(ctx, accountID)
			monitor("delete_account", err, accountID)
			if err != nil {
				break
			}

			accountID, err = a.CreateAccount(ctx, 100000)
			monitor("create_account", err, accountID)

			time.Sleep(time.Millisecond * 500)
		}
	}
}

// generates a minimum of 2 transactions up to 70% of the current balance
func generateRandomTransactions(balance int64) []transactions.Transaction {
	var list []transactions.Transaction
	for i := 0; i < r.Intn(5) + 2; i++ {
		list = append(list, transactions.Transaction{
			AmountCents: r.Int63n(70000),
			Description: "Random transaction up to 700€",
		})
	}

	return list
}

func main() {
	appId = uuid.New().String()

	// Connect to one of the three nodes by passing through an HAProxy Load Balancer.
	db, err := sql.Open("postgres", "postgresql://myuser@localhost:26257/bank?sslmode=disable")
	if err != nil {
		logFatal("Error connecting to the database: ", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logAlert("Could not close db: %v", err)
		}
	}()

	statsdC = statsd.NewClient("localhost:8125",
		statsd.MaxPacketSize(1400),
		statsd.MetricPrefix("crdb."))

	defer func() {
		if err := statsdC.Close(); err != nil {
			logAlert("Could not close statsd client: %v", err)
		}
	}()

	s := rand.NewSource(time.Now().UnixNano())
	r = rand.New(s)

	accounting := transactions.NewAccounting(db)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}
	for i := 0; i < 50; i++ {
		go reads(ctx, accounting, &wg)
		go txs(ctx, accounting, &wg)
	}

	// Spawn 1 delete routine per app
	go deletes(ctx, accounting, &wg)

	<- signals
	cancel()

	logInfo("Waiting for all go routines to stop")
	wg.Wait()
	logInfo("END")
}

func monitor(name string, result error, extra ...string) {
	metric := name
	if result == nil {
		logInfo("Op %s was successful: %s", name, extra)
		statsdC.Incr("success_op", 1)
	} else {
		switch errors.Unwrap(result) {
		case transactions.ErrBeginTx:
			metric = "error.begin_tx"
		case transactions.ErrUpdateBalance:
			metric = "error.update_tx"
		case transactions.ErrInsertTx:
			metric = "error.insert_tx"
		case transactions.ErrScanBalance:
			metric = "error.scan_balance"
		case transactions.ErrScanAccount:
			metric = "error.scan_account"
		case transactions.ErrCreateAccount:
			metric = "error.create_account"
		case transactions.ErrInsufficientFunds:
			metric = "error.insufficient_funds"
		case transactions.ErrCommitTx:
			metric = "error.commit_tx"
		case transactions.ErrRollbackTx:
			metric = "error.rollback_tx"
		case transactions.ErrDeleteAccount:
			metric = "error.delete_account"
		case transactions.ErrRowsAffected:
			metric = "error.rows_affected"
		case transactions.ErrDeleteNotFound:
			metric = "error.delete_not_found"
		default:
			metric = "error.unknown"
			logAlert("Unknown error detected in op %s: %v", name, result)
		}

		logError("Op %s return an error. metric: %s. Error: %v", name, metric, result)
		statsdC.Incr(metric, 1)
	}
}

func logEntry(level, logEntry string, args ...interface{}) {
	log.Printf(fmt.Sprintf("[%s][%s] ", appId, level) + logEntry, args...)
}

func logError(error string, args ...interface{}) {
	logEntry("ERROR", error, args...)
}

func logInfo(error string, args ...interface{}) {
	logEntry("INFO ", error, args...)
}

func logAlert(error string, args ...interface{}) {
	logEntry("ALERT", error, args...)
}

func logFatal(error string, args ...interface{}) {
	logEntry("ALERT", error, args...)
	os.Exit(-1)
}