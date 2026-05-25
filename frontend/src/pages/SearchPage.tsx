import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { fetchSearch, fetchHotBooks } from '../api';
import { useAuth } from '../contexts/AuthContext';
import { useTheme } from '../contexts/ThemeContext';
import { useAsync } from '../hooks/useAsync';
import StatusMessage from '../components/StatusMessage';
import BookCard from '../components/BookCard';
import type { Book } from '../types';

const SOURCES = [
  { key: 'guangyu', name: '番茄 (API)' },
  { key: 'biquge900', name: '笔趣阁' },
  { key: 'qixinge', name: '七星阁' },
] as const;

export default function SearchPage() {
  const [keyword, setKeyword] = useState('');
  const [searched, setSearched] = useState(false);
  const [currentSource, setCurrentSource] = useState<string>('guangyu');
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const { theme, toggleTheme } = useTheme();

  const hotBooks = useAsync<Book[]>(fetchHotBooks, { immediate: true });
  const search = useAsync<Book[]>();

  async function handleSearch() {
    const kw = keyword.trim();
    if (!kw) return;
    setSearched(true);
    await search.execute(() => fetchSearch(kw, currentSource));
    if (search.data?.length === 0) search.setError('未找到相关书籍');
  }

  function handleBackToHot() {
    setSearched(false);
    search.setData(null);
    search.setError('');
    setKeyword('');
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'Enter') handleSearch();
  }

  return (
    <div className="page search-page">
      <header className="header">
        <h1>阅读器</h1>
        <div className="header-actions">
          <button className="header-btn" onClick={toggleTheme} title="切换主题">
            {theme === 'light' ? '☀' : '☾'}
          </button>
          {user ? (
            <>
              <span style={{ fontSize: 13, color: 'var(--text-secondary)', alignSelf: 'center' }}>
                {user.username}
              </span>
              <button className="header-btn" onClick={() => navigate('/bookshelf')}>
                书架
              </button>
              <button className="header-btn" onClick={logout}>
                退出
              </button>
            </>
          ) : (
            <button className="header-btn" onClick={() => navigate('/login')}>
              登录
            </button>
          )}
        </div>
      </header>

      <div className="source-selector">
        {SOURCES.map((s) => (
          <button
            key={s.key}
            className={`source-tag ${currentSource === s.key ? 'active' : ''}`}
            onClick={() => setCurrentSource(s.key)}
          >
            {s.name}
          </button>
        ))}
      </div>

      <div className="search-bar">
        <input
          type="text"
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="输入书名搜索..."
          autoFocus
        />
        <button onClick={handleSearch} disabled={search.loading}>
          {search.loading ? '搜索中...' : '搜索'}
        </button>
      </div>

      <StatusMessage
        loading={search.loading}
        loadingSkeleton={
          <div className="book-list">
            {[1, 2, 3].map((i) => (
              <div key={i} className="skeleton-card">
                <div className="skeleton skeleton-cover" />
                <div className="skeleton-lines">
                  <div className="skeleton skeleton-line" />
                  <div className="skeleton skeleton-line" />
                  <div className="skeleton skeleton-line" />
                </div>
              </div>
            ))}
          </div>
        }
      />

      {/* 搜索结果列表 */}
      {searched && !search.loading && (
        <>
          <div className="section-header">
            <button className="link-btn" onClick={handleBackToHot}>
              ↩ 返回推荐
            </button>
            <span className="section-title">搜索结果</span>
          </div>
          {search.data && search.data.length > 0 ? (
            <div className="book-list">
              {search.data.map((book, i) => (
                <BookCard
                  key={book.bookId}
                  book={book}
                  animationDelay={i * 50}
                  onClick={() => navigate('/chapters', { state: { book } })}
                />
              ))}
            </div>
          ) : (
            <StatusMessage empty emptyText="暂无结果" />
          )}
        </>
      )}

      {/* 热门推荐（搜索前显示） */}
      {!searched && (
        <>
          <h2 className="section-title hot-title">🔥 热搜榜</h2>
          <StatusMessage
            loading={hotBooks.loading}
            loadingSkeleton={
              <div className="skeleton-grid">
                {[1, 2, 3, 4, 5, 6].map((i) => (
                  <div key={i} className="skeleton skeleton-grid-card" />
                ))}
              </div>
            }
          />
          {!hotBooks.loading && hotBooks.data && (
            <div className="hot-grid">
              {hotBooks.data.map((book, i) => (
                <div
                  key={book.bookId}
                  className="hot-card stagger-in"
                  style={{ animationDelay: `${i * 40}ms` }}
                  onClick={() => navigate('/chapters', { state: { book } })}
                >
                  <img
                    src={book.cover}
                    alt={book.title}
                    className="hot-cover"
                    loading="lazy"
                  />
                  <span className="hot-name">{book.title}</span>
                </div>
              ))}
            </div>
          )}
        </>
      )}
    </div>
  );
}
