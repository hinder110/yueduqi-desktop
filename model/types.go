package model

type Book struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	Cover       string `json:"cover"`
	Intro       string `json:"intro"`
	Kind        string `json:"kind"`
	LastChapter string `json:"lastChapter"`
	WordCount   string `json:"wordCount"`
	BookID      string `json:"bookId"`
	SourceKey   string `json:"sourceKey"`
	Source      string `json:"source"`
	Tab         string `json:"tab"`
}

type Chapter struct {
	Title  string `json:"title"`
	ItemID string `json:"itemId"`
}

type ChapterContent struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type APIResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

type SearchParams struct {
	Keyword     string
	Source      string
	InnerSource string
	InnerTab    string
}

type ChaptersParams struct {
	Source      string
	BookID      string
	InnerSource string
	InnerTab    string
}

type ContentParams struct {
	Source      string
	BookID      string
	ItemID      string
	InnerSource string
	InnerTab    string
}
