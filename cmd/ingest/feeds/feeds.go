package feeds

import (
	"encoding/json"
	"fmt"
	"os"
)

type FeedConfig struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Format string `json:"format"`
}

type FeedTopic struct {
	ID    string       `json:"id"`
	Label string       `json:"label"`
	Feeds []FeedConfig `json:"feeds"`
}

type FeedsFile struct {
	Topics []FeedTopic `json:"topics"`
}

type FeedEntry struct {
	TopicID    string
	TopicLabel string
	Feed       FeedConfig
}

func Load(path string) ([]FeedEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading feeds file: %w", err)
	}

	var f FeedsFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parsing feeds file: %w", err)
	}

	var entries []FeedEntry
	for _, topic := range f.Topics {
		for _, feed := range topic.Feeds {
			entries = append(entries, FeedEntry{
				TopicID:    topic.ID,
				TopicLabel: topic.Label,
				Feed:       feed,
			})
		}
	}
	return entries, nil
}
