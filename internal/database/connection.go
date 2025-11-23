package database

import (
	"database/sql"
	"fmt"
)

type Connection interface {
	Querier
	BeginTransaction() (TransactionQuerier, error)
}

type queries struct {
	Queries
	db *sql.DB
}

func openDb(file string) (*sql.DB, error) {
	return sql.Open("sqlite", fmt.Sprintf("%s?_txlock=immediate", file))
}

func Open(file string) (Connection, error) {
	db, err := openDb(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}
	return &queries{*New(db), db}, nil
}

type TransactionQuerier interface {
	Querier
	Commit() error
	Rollback() error
}

type transactionQueries struct {
	Queries
	tx *sql.Tx
}

func (t *transactionQueries) Commit() error {
	return t.tx.Commit()
}

func (t *transactionQueries) Rollback() error {
	return t.tx.Rollback()
}

func (q *queries) BeginTransaction() (TransactionQuerier, error) {
	tx, err := q.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return &transactionQueries{*q.WithTx(tx), tx}, nil
}
