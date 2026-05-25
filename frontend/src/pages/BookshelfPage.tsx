import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { fetchBookshelf, removeFromBookshelf } from '../api';
import { useAuth } from '../contexts/AuthContext';
import { useAsync } from '../hooks/useAsync';
import StatusMessage from '../components/StatusMessage';
import type { BookshelfItem } from '../types';

export default function BookshelfPage() {
  const { user } = useAuth();
  const navigate = useNavigate();
  const shelf = useAsync<BookshelfItem[]>(fetchBookshelf, { immediate: true });

  useEffect(() => {
    if (!user) navigate('/login', { replace: true });
  }, [user, navigate]);

  async function handleRemove(bookId: number) {
    const res = await removeFromBookshelf(bookId);
    if (res.success && shelf.data) {
      shelf.setData(shelf.data.filter((b) => b.id !== bookId));
    }
  }

  function handleOpen(book: BookshelfItem) {
    navigate('/chapters', {
      state: {
        book: {
          bookId: book.bookId,
          title: book.title,
          author: book.author,
          cover: book.cover,
          intro: book.intro,
          sourceKey: book.sourceKey,
          source: '番茄',
          tab: '小说',
        },
      },
    });
  }

  if (!user) return null;

  return (
    <div className="page bookshelf-page">
      <header className="header">
        <button className="back-btn" onClick={() => navigate(-1)}>← 返回</button>
        <h1>我的书架</h1>
      </header>

      <StatusMessage
        loading={shelf.loading}
        error={shelf.error}
        empty={!shelf.loading && (shelf.data?.length ?? 0) === 0}
        emptyText="书架空空如也，去搜索页面添加书籍吧"
      />

      <div className="bookshelf-grid">
        {shelf.data?.map((book) => (
          <div key={book.id} className="bookshelf-card">
            <button
              className="bookshelf-remove-btn"
              onClick={(e) => { e.stopPropagation(); handleRemove(book.id); }}
            >
              ✕
            </button>
            <div className="bookshelf-card-inner" onClick={() => handleOpen(book)}>
              {book.cover ? (
                <img src={book.cover} alt={book.title} className="bookshelf-card-cover" />
              ) : (
                <div className="bookshelf-card-cover bookshelf-card-cover--placeholder" />
              )}
              <h3 className="bookshelf-card-title">{book.title}</h3>
              {book.author && (
                <span className="bookshelf-card-author">{book.author}</span>
              )}
              {book.chapterIndex > 0 && (
                <span className="progress-hint">已读至第 {book.chapterIndex} 章</span>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
