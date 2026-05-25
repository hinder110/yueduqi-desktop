package main

import (
	"context"

	"yueduqi-desktop/model"
	"yueduqi-desktop/parser"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) SearchBooks(keyword string) ([]model.Book, error) {
	return parser.ForSource("guangyu").SearchBooks(a.ctx, keyword)
}

func (a *App) GetHotBooks() ([]model.Book, error) {
	return parser.GetHotBooks(a.ctx)
}

func (a *App) GetChapters(bookID string) ([]model.Chapter, error) {
	return parser.ForSource("guangyu").GetChapters(a.ctx, bookID, "番茄", "小说")
}

func (a *App) GetChapterContent(bookID, itemID string) (model.ChapterContent, error) {
	return parser.ForSource("guangyu").GetChapterContent(a.ctx, bookID, itemID, "番茄", "小说")
}
