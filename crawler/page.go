package main

import "bytes"
import "code.google.com/p/go.net/html"
import "container/list"
import "database/sql"
import "fmt"
import "regexp"
import "strconv"
import "strings"

type page struct {
	id    int
	url   string
	title string
	text  string
}

type pageWithLinks struct {
	page  *page
	links []link
}

func parseHtml(pageUrl, contents string, getAllLinks bool) *pageWithLinks {
	fmt.Println("   parsing:", pageUrl)

	inBody := false
	inScript := false
	stop := false
	href := ""
	inPageTitle := 0
	inWikiText := 0
	pageTextBuffer := new(bytes.Buffer)
	pageTitleTextBuffer := new(bytes.Buffer)
	linkTextBuffer := new(bytes.Buffer)
	linkList := list.New()
	tokenizer := html.NewTokenizer(strings.NewReader(contents))

	for !stop {
		token := tokenizer.Next()
		switch token {
		case html.ErrorToken:
			stop = true
		case html.StartTagToken:
			if inPageTitle > 0 {
				inPageTitle++
			}

			if inWikiText > 0 {
				inWikiText++
			}

			tagName, hasAttr := tokenizer.TagName()
			if spellsBody(tagName) {
				inBody = true
			} else if spells(tagName, "script") {
				inScript = true
			} else if hasAttr {
				for {
					key, val, more := tokenizer.TagAttr()
					if (getAllLinks || inWikiText > 0) && spellsA(tagName) {
						if spellsHref(key) {
							href = toUtf8(val)
							linkTextBuffer = new(bytes.Buffer)
							break
						}
					} else if spells(tagName, "div") {
						if spells(key, "class") && spells(val, "pagetitle") {
							inPageTitle = 1
						} else if spells(key, "id") && spells(val, "wikitext") {
							inWikiText = 1
						}
					}

					if !more {
						break
					}
				}
			}
		case html.TextToken:
			if inBody {
				tagText := tokenizer.Text()

				if inWikiText > 0 && !inScript {
					pageTextBuffer.Write(tagText)
				}

				if inPageTitle > 0 {
					pageTitleTextBuffer.Write(tagText)
				}

				if len(href) > 0 {
					linkTextBuffer.Write(tagText)
				}
			}
		case html.EndTagToken:
			tagName, _ := tokenizer.TagName()
			if spellsBody(tagName) {
				inBody = false
			} else if spells(tagName, "script") {
				inScript = false
			} else if spellsA(tagName) && len(href) > 0 && !strings.Contains(href, "?action=") {
				aLink := &link{
					id:       -1,
					href:     href,
					text:     strings.TrimSpace(toUtf8(linkTextBuffer.Bytes())),
					pageFrom: -1,
					pageTo:   -1,
				}
				if isContentLink(aLink) {
					linkList.PushBack(aLink)
				}
				href = ""
				linkTextBuffer = new(bytes.Buffer)
			}

			if inPageTitle > 0 {
				inPageTitle--
			}

			if inWikiText > 0 {
				inWikiText--
			}
		}
	}

	parsedPage := &page{
		id:    -1,
		url:   pageUrl,
		title: stripExtraWhitespace(pageTitleTextBuffer.Bytes()),
		text:  stripExtraWhitespace(pageTextBuffer.Bytes()),
	}

	links := listToLinks(linkList)

	return &pageWithLinks{
		page:  parsedPage,
		links: links,
	}
}

var extraWhitespaceRegex = regexp.MustCompile("\\s\\s+")

func stripExtraWhitespace(input []byte) string {
	return strings.TrimSpace(toUtf8(extraWhitespaceRegex.ReplaceAll(input, []byte{' '})))
}

func spells(bytes []byte, what string) bool {
	if len(bytes) != len(what) {
		return false
	}

	for i := 0; i < len(what); i++ {
		if bytes[i] != what[i] {
			return false
		}
	}

	return true
}

func spellsA(bytes []byte) bool {
	return spells(bytes, "a")
}

func spellsHref(bytes []byte) bool {
	return spells(bytes, "href")
}

func spellsBody(bytes []byte) bool {
	return spells(bytes, "body")
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

func toUtf8(iso8859_1_buf []byte) string {
	buf := make([]rune, len(iso8859_1_buf))
	for i, b := range iso8859_1_buf {
		buf[i] = rune(b)
	}
	return string(buf)
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

	const query = "insert into page (url, title, \"text\") values ($1, $2, $3) returning id"
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(strings.TrimPrefix(page.url, tvTroperUrlPrefix), page.title, page.text)
	if err != nil {
		return err
	}
	defer rows.Close()
	rows.Next()

	var id int64
	rows.Scan(&id)

	page.id = int(id)

	return nil
}

func (page *page) find(tx *sql.Tx) (int, error) {
	const query = "select id from page where url = $1"
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
