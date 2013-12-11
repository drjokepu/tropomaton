package main

import "database/sql"
import "fmt"
import "github.com/PuerkitoBio/goquery"
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

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(contents))
	if err != nil {
		panic(err)
	}

	title := extractTitleFromDocument(doc)
	text := extractTextFromDocument(doc)

	parsedPage := &page{
		id:    -1,
		url:   pageUrl,
		title: title,
		text:  text,
	}

	links := extractLinksFromDocument(doc, getAllLinks)

	return &pageWithLinks{
		page:  parsedPage,
		links: links,
	}
}

func extractTitleFromDocument(doc *goquery.Document) string {
	var title string
	pageTitleDiv := doc.Find("div.pagetitle")
	if pageTitleDiv.ChildrenFiltered(":not(span)").Length() > 0 {
		title = pageTitleDiv.ChildrenFiltered("span").Text()
	} else {
		title = pageTitleDiv.Text()
	}

	return stripExtraWhitespace(title)
}

func extractTextFromDocument(doc *goquery.Document) string {
	return stripExtraWhitespace(doc.Find("#wikitext").Text())
}

func extractLinksFromDocument(doc *goquery.Document, getAllLinks bool) []link {
	var linkElements *goquery.Selection
	if getAllLinks {
		linkElements = doc.Find("a")
	} else {
		linkElements = doc.Find("#wikitext a")
	}

	links := make([]link, 0)
	linkElements.Each(func(index int, linkElement *goquery.Selection) {
		href, exists := linkElement.Attr("href")
		if !exists {
			return
		}

		link := link{
			id:       -1,
			href:     strings.TrimSpace(toUtf8(href)),
			text:     stripExtraWhitespace(linkElement.Text()),
			pageFrom: -1,
			pageTo:   -1,
		}

		if link.isContentLink() {
			links = append(links, link)
		}
	})

	return links
}

var extraWhitespaceRegex = regexp.MustCompile("[\\s]{2,}")

func stripExtraWhitespace(input string) string {
	return strings.TrimSpace(extraWhitespaceRegex.ReplaceAllString(toUtf8(input), " "))
}

func toUtf8(iso8859_1_str string) string {
	iso8859_1_buf := []byte(iso8859_1_str)
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
