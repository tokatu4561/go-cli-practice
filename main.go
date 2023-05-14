package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Entry struct {
	AuthorID string
	Author string
	TitleID string
	Title string
	InfoURL string
	ZipURL string
}

// 作者とzipファイルのURLを取得
func findAuthorAndZipUrl(url string) (string, string) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return "", ""
	}

	author := doc.Find("table[summary=作家データ tr:nth-child(1) td:nth-child(2)]").Text()

	zipUrl := ""
	doc.Find("table.download a").Each(func(i int, elm *goquery.Selection) {
		href := elm.AttrOr("href", "")
		if strings.HasSuffix(href, ".zip") {
			zipUrl = href
		}
	})

	return author, zipUrl
}

func findEntries(siteURL string) ([]Entry, error) {
	// TODO
	doc, err := goquery.NewDocument(siteURL)
	if err != nil {
		return nil, err
	}

	pat := regexp.MustCompile(`.*/cards/([0-9]+)/card([0-9]+).html$`)
	doc.Find("ol li a").Each(func(i int, elem *goquery.Selection) {
		matchSli := pat.FindStringSubmatch(elem.AttrOr("href", ""))
		if len(matchSli) != 3 {
			return
		}
		pageUrl := fmt.Sprintf("https://www.aozora.gr.jp/cards/%s/card%s.html", matchSli[1], matchSli[2])

		author, zipUrl := findAuthorAndZipUrl(pageUrl)
		println(author, zipUrl)
	})

	return nil, nil
}

func main() {
	listURL := "https://www.aozora.gr.jp/index_pages/person879.html"

	entries, err := findEntries(listURL)
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range entries {
		fmt.Println(entry.Title, entry.ZipURL)
	}
}