package main

import "database/sql"
import _ "github.com/lib/pq"

func run(f func(*sql.Tx) error) error {
	db, err := sql.Open("postgres", sharedConfig.databaseConnectionString)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	err = f(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	return err
}
