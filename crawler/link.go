package main

import "database/sql"
import "strings"

type link struct {
	id       int
	href     string
	text     string
	pageFrom int
	pageTo   int
}

func (link *link) isContentLink() bool {
	const filterPrefix = "Main/"
	return (strings.HasPrefix(link.href, filterPrefix) ||
		strings.HasPrefix(link.href, tvTroperUrlPrefix+filterPrefix)) &&
		!strings.Contains(link.href, "?action=")
}

func (link *link) insert(tx *sql.Tx) error {
	const query = "insert into link (href, \"text\", page_from, page_to) values ($1, $2, $3, $4) returning id"
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	var pageToParam sql.NullInt64
	if link.pageTo < 0 {
		pageToParam = sql.NullInt64{Int64: 0, Valid: false}
	} else {
		pageToParam = sql.NullInt64{Int64: int64(link.pageTo), Valid: true}
	}

	rows, err := stmt.Query(strings.TrimPrefix(link.href, tvTroperUrlPrefix), link.text, link.pageFrom, pageToParam)
	if err != nil {
		return err
	}
	defer rows.Close()
	rows.Next()

	var id int
	rows.Scan(&id)

	link.id = int(id)

	return nil
}

func (link *link) findPage(tx *sql.Tx) error {
	const query = "select id from page where url = $1"
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(strings.TrimPrefix(link.href, tvTroperUrlPrefix))
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil
	}

	var id int

	rows.Scan(&id)
	link.pageTo = id

	return nil
}

func getNextLink() (*link, error) {
	var nextLink *link
	err := run(func(tx *sql.Tx) error {
		nextLinkInner, err := selectNextLink(tx)
		nextLink = nextLinkInner
		return err
	})
	return nextLink, err
}

func selectNextLink(tx *sql.Tx) (*link, error) {
	const query = "select id, href, page_from from link where page_to is null limit 1"
	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var id, pageFrom int
	var href string

	rows.Scan(&id, &href, &pageFrom)
	return &link{
		id:       id,
		href:     href,
		text:     "",
		pageFrom: pageFrom,
		pageTo:   -1,
	}, nil
}

func updateIncomingLinksWithPageId(href string, pageId int, tx *sql.Tx) error {
	const query = "update link set page_to = $1 where href = $2"

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(pageId, strings.TrimPrefix(href, tvTroperUrlPrefix))
	return err
}
