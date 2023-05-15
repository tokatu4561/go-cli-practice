package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Entry struct {
	AuthorID string
	Author string
	TitleID string
	Title string
	SiteURL string
	ZipURL string
}

// 作者とzipファイルのURLを取得
func findAuthorAndZipUrl(siteUrl string) (string, string) {
	doc, err := goquery.NewDocument(siteUrl)
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

	if zipUrl == "" {
		return author, ""
	}

	u, err := url.Parse(siteUrl)
	if err != nil {
		return author, ""
	}

	u.Path = path.Join(path.Dir(u.Path), zipUrl)

	return author, u.String()
}

func findEntries(siteURL string) ([]Entry, error) {
	// TODO
	doc, err := goquery.NewDocument(siteURL)
	if err != nil {
		return nil, err
	}

	entries := []Entry{}
	pat := regexp.MustCompile(`.*/cards/([0-9]+)/card([0-9]+).html$`)
	doc.Find("ol li a").Each(func(i int, elem *goquery.Selection) {
		matchSli := pat.FindStringSubmatch(elem.AttrOr("href", ""))
		if len(matchSli) != 3 {
			return
		}

		title := elem.Text()
		authorId := matchSli[1]
		titleId := matchSli[2]
		pageUrl := fmt.Sprintf("https://www.aozora.gr.jp/cards/%s/card%s.html", authorId, titleId)

		author, zipUrl := findAuthorAndZipUrl(pageUrl)
		if zipUrl != "" {
			entries = append(entries, Entry{
				AuthorID: authorId,
				Author: author,
				TitleID: titleId,
				Title: title,
				SiteURL: siteURL,
				ZipURL: zipUrl,
			})
		}
	})

	return entries, nil
}

func extractText(zipUrl string) (string, error) {
	res, err := http.Get(zipUrl)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return "", err
	}

	for _, file := range r.File {
		if path.Ext(file.Name) == ".txt" {
			f, err := file.Open()
			if err != nil {
				return "", err
			}
			b, err := ioutil.ReadAll(f)
			f.Close()
			if err != nil {
				return "", err
			}
			return string(b), nil
		}
	}

	return "", nil
}

func main() {
	listURL := "https://www.aozora.gr.jp/index_pages/person879.html"

	entries, err := findEntries(listURL)
	if err != nil {
		log.Fatal(err)
	}
	for _, entry := range entries {
		fmt.Println(entry.Title, entry.ZipURL)
		content, err := extractText(entry.ZipURL)
		if err != nil {
			log.Fatal(err)
		}
		println(content)
	}
}