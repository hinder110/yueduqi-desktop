export interface Book {
  title: string;
  author?: string;
  cover?: string;
  intro?: string;
  kind?: string;
  lastChapter?: string;
  wordCount?: string;
  bookId: string;
  /** 书源标识: guangyu | biquge900 */
  sourceKey: string;
  /** 光遇内部书源：番茄/七猫等 */
  source: string;
  tab: string;
}

export interface Chapter {
  title: string;
  itemId: string;
}

export interface ChapterContent {
  title: string;
  content: string;
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
}

export interface User {
  id: string;
  username: string;
}

export interface LoginData {
  token: string;
  user: User;
}

export interface BookshelfItem {
  id: number;
  title: string;
  author: string;
  cover: string;
  intro: string;
  bookId: string;
  sourceKey: string;
  addedAt: string;
  chapterIndex: number;
  chapterItemId: string;
}
