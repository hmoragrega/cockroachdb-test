# CockroachDB test

This test consists in bringing up a cluster of 3 nodes with a HAProxy Load Balancer in front of them.

Once the cluster is up and running a Golang application will fire 2000 transactions over 10 connections that all try
to write over the same records thus over the same Range.

Every transaction consists of two `INSERT`s, one `UPDATE` and one `SELECT`.

The concept is very simple, an account is created in the `accounts` table with a balance of 1,000.00€ then 2000
goroutines are created. Each one of them tries a Database transaction. 

Each Database transaction tries to add two monetary transactions to the `transactions` table for the exact same account, 
one of 500.00€ and one of 495.00€.  

After writing the two monetary transactions the Database transaction tries to update the account balance. 
It then `SELECT`s the balance to check whether it is below zero. 
If it is below zero it rollbacks the transaction, if it is not below zero it commits the Database transaction.

We expect only one Database transaction to succeed out of the 2000, meaning the remaining balance at the end of the
program should be 5.00€.

# Run the example

1. Bring the cluster up with `make run`
2. Once the cluster is up and running then run the test with `make test` 
