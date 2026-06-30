package parser

import (
	"context"

	"yueduqi-desktop/model"
)

type Parser interface {
	SearchBooks(ctx context.Context, keyword string) ([]model.Book, error)
	GetChapters(ctx context.Context, bookID, innerSource, innerTab string) ([]model.Chapter, error)
	GetChapterContent(ctx context.Context, bookID, itemID, innerSource, innerTab string) (model.ChapterContent, error)
}

// parsers is populated by Register() calls, typically from init() in each parser file.
var parsers = map[string]Parser{}

// Register adds a Parser for the given source key. Callers should register
// before ForSource is invoked (e.g. in an init function).
func Register(source string, p Parser) {
	parsers[source] = p
}

// ForSource returns the registered Parser for source, falling back to
// GuangyuParser when no match is found so that unrecognised sources
// still receive results from the broadest aggregator.
func ForSource(source string) Parser {
	if p, ok := parsers[source]; ok {
		return p
	}
	// catch-all fallback: guangyu aggregates many upstream sources
	return &GuangyuParser{}
}
