// @vitest-environment jsdom

import { describe, it, expect, afterEach, vi } from 'vitest';
import { render, screen, cleanup } from '@testing-library/react';
import '@testing-library/jest-dom/vitest';
import ErrorBoundary from './ErrorBoundary';

const store = new Map<string, string>();
vi.stubGlobal('localStorage', {
  getItem: vi.fn((key: string) => store.get(key) ?? null),
  setItem: vi.fn((key: string, value: string) => { store.set(key, value); }),
  removeItem: vi.fn((key: string) => { store.delete(key); }),
  clear: vi.fn(() => { store.clear(); }),
});

function Explode(): React.ReactElement {
  throw new Error('模拟崩溃');
}

function Normal() {
  return <span>正常内容</span>;
}

describe('ErrorBoundary — 错误边界', () => {
  afterEach(() => {
    cleanup();
  });

  it('正常渲染子组件', () => {
    render(
      <ErrorBoundary>
        <Normal />
      </ErrorBoundary>
    );
    expect(screen.getByText('正常内容')).toBeInTheDocument();
  });

  it('子组件崩溃时显示错误提示', () => {
    // 压制 React 的 error log
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {});
    render(
      <ErrorBoundary>
        <Explode />
      </ErrorBoundary>
    );
    expect(screen.getByText(/意外错误/)).toBeInTheDocument();
    spy.mockRestore();
  });

  it('崩溃后显示返回首页按钮', () => {
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {});
    render(
      <ErrorBoundary>
        <Explode />
      </ErrorBoundary>
    );
    expect(screen.getByText('返回首页')).toBeInTheDocument();
    spy.mockRestore();
  });
});
