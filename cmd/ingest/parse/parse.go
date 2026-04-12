package parse

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/mmcdole/gofeed"
)

type Item struct {
	Title       string
	Link        string
	GUID        string
	PubDate     string
	Content     string
	Description string
}

func Parse(xmlStr, format string) ([]Item, error) {
	if format == "rdf" {
		return parseRDF(xmlStr)
	}
	return parseRSS(xmlStr)
}

func parseRSS(xmlStr string) ([]Item, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseString(xmlStr)
	if err != nil {
		return nil, fmt.Errorf("parsing RSS: %w", err)
	}

	items := make([]Item, 0, len(feed.Items))
	for _, fi := range feed.Items {
		item := Item{
			Title: fi.Title,
			Link:  fi.Link,
			GUID:  fi.GUID,
		}

		// Date: prefer PublishedParsed formatted as RFC3339
		if fi.PublishedParsed != nil {
			item.PubDate = fi.PublishedParsed.Format("2006-01-02T15:04:05Z07:00")
		} else if fi.UpdatedParsed != nil {
			item.PubDate = fi.UpdatedParsed.Format("2006-01-02T15:04:05Z07:00")
		} else {
			item.PubDate = fi.Published
		}

		// Content: content:encoded > description > content snippet
		if v, ok := fi.Extensions["content"]["encoded"]; ok && len(v) > 0 {
			item.Content = v[0].Value
		}
		item.Description = fi.Description
		if item.Description == "" {
			item.Description = fi.Content
		}

		items = append(items, item)
	}
	return items, nil
}

// rdfDoc represents the top-level RDF document
type rdfDoc struct {
	XMLName xml.Name  `xml:"RDF"`
	Items   []rdfItem `xml:"item"`
}

type rdfItem struct {
	About       string `xml:"http://www.w3.org/1999/02/22-rdf-syntax-ns# about,attr"`
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Date        string `xml:"http://purl.org/dc/elements/1.1/ date"`
	GUID        string `xml:"guid"`
}

func parseRDF(xmlStr string) ([]Item, error) {
	// Preprocess: collapse CDATA and ensure proper namespace handling
	decoder := xml.NewDecoder(strings.NewReader(xmlStr))

	var doc rdfDoc
	if err := decoder.Decode(&doc); err != nil {
		return nil, fmt.Errorf("parsing RDF: %w", err)
	}

	items := make([]Item, 0, len(doc.Items))
	for _, ri := range doc.Items {
		guid := ri.GUID
		if guid == "" {
			guid = ri.About
		}
		if guid == "" {
			guid = ri.Link
		}

		items = append(items, Item{
			Title:       ri.Title,
			Link:        ri.Link,
			GUID:        guid,
			PubDate:     ri.Date,
			Description: ri.Description,
		})
	}
	return items, nil
}
