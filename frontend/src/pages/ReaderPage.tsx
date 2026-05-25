import { useEffect, useState, useCallback, useRef } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { fetchContent, fetchBookshelf, updateProgress } from '../api';
import { useAuth } from '../contexts/AuthContext';
import StatusMessage from '../components/StatusMessage';
import type { Book, Chapter, ChapterContent } from '../types';

type Theme = 'light' | 'dark';
type FontSize = 'sm' | 'md' | 'lg';

const FONT_LABELS: Record<FontSize, string> = { sm: 'A⁻', md: 'A', lg: 'A⁺' };
const FONT_NEXT: Record<FontSize, FontSize> = { sm: 'md', md: 'lg', lg: 'sm' };

export default function ReaderPage() {
  const location = useLocation();
  const navigate = useNavigate();
  const { user } = useAuth();

  const book = location.state?.book as Book | undefined;
  const initialChapter = location.state?.chapter as Chapter | undefined;
  const chapters = location.state?.chapters as Chapter[] | undefined;
  const initialIndex = (location.state?.currentIndex as number) ?? 0;

  const [currentIndex, setCurrentIndex] = useState(initialIndex);
  const [content, setContent] = useState<ChapterContent | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [theme, setTheme] = useState<Theme>('light');
  const [fontSize, setFontSize] = useState<FontSize>('md');
  const preloaded = useRef<Map<string, ChapterContent>>(new Map());
  const shelfBookId = useRef<number | null>(null);
  const progressSynced = useRef(false);

  const currentChapter = chapters ? chapters[currentIndex] : initialChapter;

  const loadContent = useCallback(
    async (ch: Chapter) => {
      if (!book) return;
      setLoading(true);
      setError('');

      // 检查是否已预加载
      const cached = preloaded.current.get(ch.itemId);
      if (cached) {
        setContent(cached);
        preloaded.current.delete(ch.itemId);
        setLoading(false);
        return;
      }

      try {
        const res = await fetchContent(book.bookId, ch.itemId, book.sourceKey, book.source, book.tab);
        if (res.success && res.data) {
          setContent(res.data);
        } else {
          setError(res.error ?? '加载正文失败');
        }
      } catch {
        setError('请求异常，请稍后重试');
      } finally {
        setLoading(false);
      }
    },
    [book]
  );

  // 预加载下一章
  const preloadNextChapter = useCallback(
    async (index: number) => {
      if (!book || !chapters) return;
      const nextIndex = index + 1;
      if (nextIndex >= chapters.length) return;
      const nextCh = chapters[nextIndex];
      if (preloaded.current.has(nextCh.itemId)) return;
      try {
        const res = await fetchContent(book.bookId, nextCh.itemId, book.sourceKey, book.source, book.tab);
        if (res.success && res.data) {
          preloaded.current.set(nextCh.itemId, res.data);
        }
      } catch {
        // 预加载失败不影响当前阅读
      }
    },
    [book, chapters]
  );

  // 查找书架条目并同步阅读进度
  const syncProgress = useCallback(
    async (index: number) => {
      if (!user || !book || !chapters) return;
      const ch = chapters[index];
      if (!ch) return;
      try {
        if (shelfBookId.current === null) {
          const res = await fetchBookshelf();
          if (res.success && res.data) {
            const found = res.data.find((item: any) => item.bookId === book.bookId);
            if (found) shelfBookId.current = found.id;
          }
        }
        if (shelfBookId.current !== null) {
          await updateProgress(shelfBookId.current, index, ch.itemId);
        }
      } catch {
        // 进度同步失败不影响阅读
      }
    },
    [user, book, chapters]
  );

  useEffect(() => {
    if (!book || !currentChapter) {
      navigate('/', { replace: true });
      return;
    }
    loadContent(currentChapter);
  }, [currentChapter]);

  // 内容加载完成后：同步进度 + 预加载下一章
  useEffect(() => {
    if (!content || !book) return;
    if (!progressSynced.current) {
      syncProgress(currentIndex);
      progressSynced.current = true;
    }
    preloadNextChapter(currentIndex);
  }, [content]);

  function goToChapter(index: number) {
    if (!chapters || index < 0 || index >= chapters.length) return;
    setCurrentIndex(index);
    setContent(null);
    window.scrollTo(0, 0);
    syncProgress(index);
  }

  if (!book || !currentChapter) return null;

  const hasPrev = chapters ? currentIndex > 0 : false;
  const hasNext = chapters ? currentIndex < chapters.length - 1 : false;

  return (
    <div className={`reader-page theme-${theme} font-${fontSize}`}>
      {/* 顶栏 */}
      <div className="reader-topbar">
        <button className="topbar-btn" onClick={() => navigate(-1)}>
          ← 返回
        </button>
        <span className="topbar-title">{book.title}</span>
        <div className="topbar-actions">
          <button
            className="topbar-btn"
            onClick={() => setFontSize(FONT_NEXT[fontSize])}
            title="切换字号"
          >
            {FONT_LABELS[fontSize]}
          </button>
          <button
            className="topbar-btn"
            onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')}
            title={theme === 'light' ? '切换夜间模式' : '切换日间模式'}
          >
            {theme === 'light' ? '🌙' : '☀️'}
          </button>
        </div>
      </div>

      {/* 章节标题 */}
      <h2 className="chapter-heading">{currentChapter.title}</h2>

      {/* 正文 */}
      <StatusMessage loading={loading} error={error} />
      {content && (
        <div className="content-body" dangerouslySetInnerHTML={{ __html: content.content }} />
      )}

      {/* 底栏翻页 */}
      <div className="reader-nav">
        <button disabled={!hasPrev} onClick={() => goToChapter(currentIndex - 1)}>
          上一章
        </button>
        <span className="chapter-indicator">
          {currentIndex + 1} / {chapters?.length ?? 1}
        </span>
        <button disabled={!hasNext} onClick={() => goToChapter(currentIndex + 1)}>
          下一章
        </button>
      </div>
    </div>
  );
}
