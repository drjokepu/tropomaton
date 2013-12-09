package main

import "bytes"
import "code.google.com/p/go.net/html"
import "container/list"
import "database/sql"
import "fmt"
import "strconv"
import "strings"

type page struct {
	id   int
	url  string
	text string
}

type pageWithLinks struct {
	page  *page
	links []link
}

func parseHtml(pageUrl, contents string) *pageWithLinks {
	fmt.Println("   parsing:", pageUrl)

	inBody := false
	stop := false
	href := ""
	pageTextBuffer := new(bytes.Buffer)
	linkTextBuffer := new(bytes.Buffer)
	linkList := list.New()
	tokenizer := html.NewTokenizer(strings.NewReader(contents))

	for !stop {
		token := tokenizer.Next()
		switch token {
		case html.ErrorToken:
			stop = true
		case html.StartTagToken:
			tagName, hasAttr := tokenizer.TagName()
			if spellsBody(tagName) {
				inBody = true
			} else if spellsA(tagName) && hasAttr {
				for {
					key, val, more := tokenizer.TagAttr()
					if spellsHref(key) {
						href = string(val)
						linkTextBuffer = new(bytes.Buffer)
						break
					} else {
						if !more {
							break
						}
					}
				}
			}
		case html.TextToken:
			if inBody {
				tagText := tokenizer.Text()
				pageTextBuffer.Write(tagText)

				if len(href) > 0 {
					linkTextBuffer.Write(tagText)
				}
			}
		case html.EndTagToken:
			tagName, _ := tokenizer.TagName()
			if spellsBody(tagName) {
				inBody = false
			} else if spellsA(tagName) && len(href) > 0 {
				aLink := &link{
					id:       -1,
					href:     href,
					text:     strings.TrimSpace(linkTextBuffer.String()),
					pageFrom: -1,
					pageTo:   -1,
				}
				if isContentLink(aLink) {
					linkList.PushBack(aLink)
				}
				href = ""
				linkTextBuffer = new(bytes.Buffer)
			}
		}
	}

	parsedPage := &page{
		id:   -1,
		url:  pageUrl,
		text: strings.TrimSpace(pageTextBuffer.String()),
	}

	links := listToLinks(linkList)

	return &pageWithLinks{
		page:  parsedPage,
		links: links,
	}
}

func spellsA(bytes []byte) bool {
	return len(bytes) == 1 && (bytes[0] == 'a' || bytes[0] == 'A')
}

func spellsHref(bytes []byte) bool {
	return len(bytes) == 4 && bytes[0] == 'h' && bytes[1] == 'r' && bytes[2] == 'e' && bytes[3] == 'f'
}

func spellsBody(bytes []byte) bool {
	return len(bytes) == 4 && bytes[0] == 'b' && bytes[1] == 'o' && bytes[2] == 'd' && bytes[3] == 'y'
}

func listToLinks(linkList *list.List) []link {
	linksArray := make([]link, linkList.Len())
	cursor := 0
	for e := linkList.Front(); e != nil; e = e.Next() {
		linksArray[cursor] = *e.Value.(*link)
		cursor++
	}
	return linksArray
}

func (pl *pageWithLinks) save() error {
	fmt.Println("    saving:", pl.page.url, "("+strconv.Itoa(len(pl.links))+" outlinks)")

	err := run(func(tx *sql.Tx) error {
		err := pl.page.insert(tx)
		if err != nil {
			return err
		}

		for _, pageLink := range pl.links {
			pageLink.pageFrom = pl.page.id

			err = pageLink.findPage(tx)
			if err != nil {
				return err
			}

			err = pageLink.insert(tx)
			if err != nil {
				return err
			}
		}

		err = updateIncomingLinksWithPageId(pl.page.url, pl.page.id, tx)
		return err
	})

	return err
}

func (page *page) insert(tx *sql.Tx) error {
	pageId, err := page.find(tx)
	if err != nil {
		return err
	}
	if pageId >= 0 {
		page.id = pageId
		return nil
	}

	const query = "insert into page (url, \"text\") values (?, ?)"
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(strings.TrimPrefix(page.url, tvTroperUrlPrefix), page.text)
	if err != nil {
		return err
	}

	id64, err := res.LastInsertId()
	if err != nil {
		return err
	}

	page.id = int(id64)

	return nil
}

func (page *page) find(tx *sql.Tx) (int, error) {
	const query = "select id from page where url = ?"
	stmt, err := tx.Prepare(query)
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(strings.TrimPrefix(page.url, tvTroperUrlPrefix))
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	if !rows.Next() {
		return -1, nil
	}

	var id int
	rows.Scan(&id)
	return id, nil
}

func hasAnyPages() (bool, error) {
	var result bool

	err := run(func(tx *sql.Tx) error {
		localResult, err := hasAnyPagesWithTx(tx)
		result = localResult
		return err
	})

	return result, err
}

func hasAnyPagesWithTx(tx *sql.Tx) (bool, error) {
	const query = "select true from page limit 1"
	stmt, err := tx.Prepare(query)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return false, err
	}
	defer rows.Close()

	return rows.Next(), nil
}
