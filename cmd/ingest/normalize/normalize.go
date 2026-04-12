package normalize

import (
	"regexp"
	"strings"
	"time"

	"github.com/tech-digest/ingest/parse"
)

var dateFormats = []string{
	time.RFC3339,
	time.RFC1123Z,
	time.RFC1123,
	"02 Jan 2006 15:04:05 -0700",
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05",
	"Mon, 2 Jan 2006 15:04:05 -0700",
	"Mon, 2 Jan 2006 15:04:05 MST",
}

var japaneseRe = regexp.MustCompile(`[\x{3040}-\x{30FF}\x{4E00}-\x{9FAF}]`)

type Item struct {
	GUID        string
	Title       string
	Link        string
	PublishedAt *time.Time
	SummaryRaw  string
}

func Normalize(item parse.Item) *Item {
	title := cleanText(item.Title)
	link := cleanText(item.Link)
	if title == "" || link == "" {
		return nil
	}

	guid := cleanText(item.GUID)
	if guid == "" {
		guid = link
	}

	publishedAt := parseDate(item.PubDate)

	summaryRaw := pickFirst(item.Content, item.Description, title)

	return &Item{
		GUID:        guid,
		Title:       title,
		Link:        link,
		PublishedAt: publishedAt,
		SummaryRaw:  cleanText(summaryRaw),
	}
}

func DetectLanguage(text string) string {
	if japaneseRe.MatchString(text) {
		return "ja"
	}
	return "unknown"
}

func cleanText(s string) string {
	if s == "" {
		return ""
	}
	return strings.Join(strings.Fields(s), " ")
}

func pickFirst(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func parseDate(value string) *time.Time {
	if value == "" {
		return nil
	}
	for _, format := range dateFormats {
		if t, err := time.Parse(format, value); err == nil {
			return &t
		}
	}
	return nil
}
