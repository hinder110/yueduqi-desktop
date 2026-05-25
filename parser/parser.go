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

func ForSource(source string) Parser {
	switch source {
	case "biquge900":
		return &BiqugeParser{}
	case "qixinge":
		return &QixingeParser{}
	default:
		return &GuangyuParser{}
	}
}
