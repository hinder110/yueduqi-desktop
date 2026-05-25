// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from 'vitest';
import { render, screen, fireEvent, cleanup } from '@testing-library/react';
import BookCard from './BookCard';
import type { Book } from '../types';

const mockBook: Book = {
  bookId: '123',
  title: '测试小说',
  author: '测试作者',
  cover: 'https://example.com/cover.jpg',
  kind: '玄幻',
  lastChapter: '第一千章 大结局',
  intro: '这是一本测试小说',
  sourceKey: 'guangyu',
  source: '番茄',
  tab: '小说',
};

describe('BookCard', () => {
  afterEach(cleanup);

  it('渲染书名和作者', () => {
    render(<BookCard book={mockBook} />);
    expect(screen.getByText('测试小说')).toBeDefined();
    expect(screen.getByText('测试作者')).toBeDefined();
  });

  it('渲染封面图片', () => {
    render(<BookCard book={mockBook} />);
    const img = screen.getByAltText('测试小说') as HTMLImageElement;
    expect(img).toBeDefined();
    expect(img.src).toBe('https://example.com/cover.jpg');
  });

  it('渲染分类', () => {
    render(<BookCard book={mockBook} />);
    expect(screen.getByText('玄幻')).toBeDefined();
  });

  it('渲染最新章节', () => {
    render(<BookCard book={mockBook} />);
    expect(screen.getByText('最新: 第一千章 大结局')).toBeDefined();
  });

  it('渲染简介', () => {
    render(<BookCard book={mockBook} />);
    expect(screen.getByText('这是一本测试小说')).toBeDefined();
  });

  it('点击触发 onClick', () => {
    const onClick = vi.fn();
    render(<BookCard book={mockBook} onClick={onClick} />);
    fireEvent.click(screen.getByText('测试小说').closest('.book-card')!);
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it('不传 onClick 时点击不报错', () => {
    render(<BookCard book={mockBook} />);
    fireEvent.click(screen.getByText('测试小说').closest('.book-card')!);
  });

  it('显示阅读进度', () => {
    render(<BookCard book={mockBook} progressChapter={42} />);
    expect(screen.getByText('上次读到第 42 章')).toBeDefined();
  });

  it('进度为 0 时不显示进度', () => {
    render(<BookCard book={mockBook} progressChapter={0} />);
    expect(screen.queryByText(/上次读到/)).toBeNull();
  });

  it('没有作者时不渲染作者元素', () => {
    const bookWithoutAuthor = { ...mockBook, author: undefined };
    const { container } = render(<BookCard book={bookWithoutAuthor} />);
    expect(container.querySelector('.book-author')).toBeNull();
  });

  it('没有封面时不渲染图片', () => {
    const bookWithoutCover = { ...mockBook, cover: undefined };
    const { container } = render(<BookCard book={bookWithoutCover} />);
    expect(container.querySelector('.book-cover')).toBeNull();
  });

  it('渲染动画延迟样式', () => {
    const { container } = render(<BookCard book={mockBook} animationDelay={150} />);
    const card = container.querySelector('.book-card') as HTMLElement;
    expect(card.style.animationDelay).toBe('150ms');
  });
});
