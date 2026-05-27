package main

import (
	"context"
	"log"
	"os"

	"yueduqi-desktop/model"
	"yueduqi-desktop/parser"
)

var logger *log.Logger

func init() {
	f, err := os.Create("/tmp/yueduqi-debug.log")
	if err != nil {
		f = os.Stderr
	}
	logger = log.New(f, "[yueduqi] ", log.LstdFlags|log.Lmsgprefix)
	logger.Println("===== app started =====")
}

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	logger.Println("startup called, ctx set")
}

func (a *App) SearchBooks(keyword string, source string) ([]model.Book, error) {
	logger.Printf("SearchBooks called: keyword=%q source=%q", keyword, source)
	books, err := parser.ForSource(source).SearchBooks(a.ctx, keyword)
	logger.Printf("SearchBooks result: %d books, err=%v", len(books), err)
	return books, err
}

func (a *App) GetHotBooks() ([]model.Book, error) {
	logger.Println("GetHotBooks called")
	books, err := parser.GetHotBooks(a.ctx)
	logger.Printf("GetHotBooks result: %d books, err=%v", len(books), err)
	return books, err
}

func (a *App) GetChapters(bookID string, source string, innerSource string, innerTab string) ([]model.Chapter, error) {
	logger.Printf("GetChapters called: bookID=%q source=%q innerSource=%q innerTab=%q", bookID, source, innerSource, innerTab)
	if innerSource == "" { innerSource = "番茄" }
	if innerTab == "" { innerTab = "小说" }
	chapters, err := parser.ForSource(source).GetChapters(a.ctx, bookID, innerSource, innerTab)
	logger.Printf("GetChapters result: %d chapters, err=%v", len(chapters), err)
	return chapters, err
}

func (a *App) GetChapterContent(bookID, itemID string, source string, innerSource string, innerTab string) (model.ChapterContent, error) {
	logger.Printf("GetChapterContent called: bookID=%q itemID=%q source=%q innerSource=%q innerTab=%q", bookID, itemID, source, innerSource, innerTab)
	if innerSource == "" { innerSource = "番茄" }
	if innerTab == "" { innerTab = "小说" }
	content, err := parser.ForSource(source).GetChapterContent(a.ctx, bookID, itemID, innerSource, innerTab)
	logger.Printf("GetChapterContent result: title=%q, len=%d, err=%v", content.Title, len(content.Content), err)
	return content, err
}
