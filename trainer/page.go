package main

import "database/sql"

type page struct {
	id    int
	url   string
	title string
	text  string
}

func getPage(pageId int, tx *sql.Tx) (*page, error) {
	const query = "select id, url, title, text from page where id = $1"
	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(pageId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var id int
	var url string
	var title string
	var text string
	rows.Scan(&id, &url, &title, &text)

	return &page{
		id:    id,
		url:   url,
		title: title,
		text:  text,
	}, nil
}
