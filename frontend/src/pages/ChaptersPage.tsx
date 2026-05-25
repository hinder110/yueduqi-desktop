import { useEffect, useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { fetchChapters, fetchBookshelf, addToBookshelf } from '../api';
import { useAuth } from '../contexts/AuthContext';
import { useAsync } from '../hooks/useAsync';
import StatusMessage from '../components/StatusMessage';
import type { Book, Chapter, BookshelfItem } from '../types';

export default function ChaptersPage() {
  const location = useLocation();
  const navigate = useNavigate();
  const book = location.state?.book as Book | undefined;
  const { user } = useAuth();

  const chapters = useAsync<Chapter[]>();
  const [addedToShelf, setAddedToShelf] = useState(false);
  const [shelfItem, setShelfItem] = useState<BookshelfItem | null>(null);

  useEffect(() => {
    if (!book) {
      navigate('/', { replace: true });
      return;
    }
    chapters.execute(() => fetchChapters(book.bookId, book.sourceKey, book.source, book.tab));
    if (user) loadShelfInfo();
  }, []);

  async function loadShelfInfo() {
    if (!book || !user) return;
    try {
      const res = await fetchBookshelf();
      if (res.success && res.data) {
        const found = res.data.find((item: any) => item.bookId === book.bookId);
        if (found) {
          setShelfItem(found);
          setAddedToShelf(true);
        }
      }
    } catch {
      // 书架信息加载失败不影响章节浏览
    }
  }

  async function handleAddToShelf() {
    if (!book || !user) return;
    const res = await addToBookshelf({
      title: book.title,
      author: book.author,
      cover: book.cover,
      intro: book.intro,
      bookId: book.bookId,
      sourceKey: book.sourceKey,
    });
    if (res.success) {
      setAddedToShelf(true);
    }
  }

  if (!book) return null;

  return (
    <div className="page chapters-page">
      <header className="header">
        <button className="back-btn" onClick={() => navigate(-1)}>
          ← 返回
        </button>
        <h1 className="header-title">{book.title}</h1>
        <span className="header-sub">{book.author}</span>
        {user && (
          <button
            className="header-btn"
            onClick={handleAddToShelf}
            disabled={addedToShelf}
          >
            {addedToShelf ? '已加入书架' : '加入书架'}
          </button>
        )}
      </header>

      {/* 书籍详情卡片 */}
      <div className="book-detail-card">
        {book.cover && (
          <img src={book.cover} alt={book.title} className="book-detail-cover" />
        )}
        <div className="book-detail-info">
          <h2 className="book-title">{book.title}</h2>
          {book.author && <span className="book-author">{book.author}</span>}
          {book.kind && <span className="book-kind">{book.kind}</span>}
          {shelfItem && shelfItem.chapterIndex > 0 && (
            <span className="progress-hint">上次读到第 {shelfItem.chapterIndex} 章</span>
          )}
          {book.intro && <p className="book-intro">{book.intro}</p>}
        </div>
      </div>

      <StatusMessage
        loading={chapters.loading}
        loadingSkeleton={
          <div className="chapter-list">
            {Array.from({ length: 8 }).map((_, i) => (
              <div key={i} className="skeleton skeleton-item" />
            ))}
          </div>
        }
        error={chapters.error}
      />

      <div className="section-header">
        <span className="section-title">目录</span>
        <span style={{ fontSize: 13, color: 'var(--text-secondary)' }}>
          共 {chapters.data?.length ?? 0} 章
        </span>
      </div>

      <div className="chapter-list">
        {chapters.data?.map((ch, i) => (
          <div
            key={ch.itemId}
            className="chapter-item stagger-in"
            style={{ animationDelay: `${i * 30}ms` }}
            onClick={() =>
              navigate('/reader', {
                state: {
                  book,
                  chapter: ch,
                  chapters: chapters.data,
                  currentIndex: i,
                },
              })
            }
          >
            <span className="chapter-index">{i + 1}</span>
            <span className="chapter-title">{ch.title}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
