package main

import "database/sql"
import "os"
import _ "github.com/mattn/go-sqlite3"

const databaseFilename = "./tropomaton.db"

func run(f func(*sql.Tx) error) error {
	db, err := sql.Open("sqlite3", databaseFilename)
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

func initdb() (bool, error) {
	if !fileExists(databaseFilename) {
		err := run(func(tx *sql.Tx) error {
			err := createSchema(tx)
			if err != nil {
				return err
			}

			return nil
		})
		return true, err
	} else {
		return false, nil
	}
}

func createSchema(tx *sql.Tx) error {
	ddl := []string{
		"create table page (" +
			"id integer not null primary key autoincrement, " +
			"url text not null, " +
			"\"text\" text not null, " +
			"constraint uk_page_url unique (url));",

		"create table link (" +
			"id integer not null primary key autoincrement, " +
			"href text not null, " +
			"page_from integer not null, " +
			"page_to integer, " +
			"\"text\" text not null);",

		"create index idx_link_href on link(href);",
	}

	for _, query := range ddl {
		err := func() error {
			stmt, err := tx.Prepare(query)
			if err != nil {
				return err
			}
			defer stmt.Close()

			_, err = stmt.Exec()
			return err
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
