import type { Book } from '../types';

interface BookCardProps {
  book: Book;
  onClick?: () => void;
  animationDelay?: number;
  /** 如果提供，显示在第 N 章之前 */
  progressChapter?: number;
}

export default function BookCard({ book, onClick, animationDelay, progressChapter }: BookCardProps) {
  return (
    <div
      className="book-card stagger-in"
      style={animationDelay !== undefined ? { animationDelay: `${animationDelay}ms` } : undefined}
      onClick={onClick}
    >
      {book.cover && <img src={book.cover} alt={book.title} className="book-cover" />}
      <div className="book-info">
        <h3 className="book-title">{book.title}</h3>
        {book.author && <span className="book-author">{book.author}</span>}
        {book.kind && <span className="book-kind">{book.kind}</span>}
        {book.lastChapter && <span className="book-last">最新: {book.lastChapter}</span>}
        {progressChapter !== undefined && progressChapter > 0 && (
          <span className="progress-hint">上次读到第 {progressChapter} 章</span>
        )}
        {book.intro && <p className="book-intro">{book.intro}</p>}
      </div>
    </div>
  );
}
