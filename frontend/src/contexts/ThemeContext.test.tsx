// @vitest-environment jsdom

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, cleanup } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import '@testing-library/jest-dom/vitest';
import { ThemeProvider, useTheme } from './ThemeContext';

// Node 22 的 jsdom 缺少全局 localStorage，手动补
const store = new Map<string, string>();
vi.stubGlobal('localStorage', {
  getItem: vi.fn((key: string) => store.get(key) ?? null),
  setItem: vi.fn((key: string, value: string) => { store.set(key, value); }),
  removeItem: vi.fn((key: string) => { store.delete(key); }),
  clear: vi.fn(() => { store.clear(); }),
});

function ShowTheme() {
  const { theme, toggleTheme } = useTheme();
  return (
    <div>
      <span data-testid="theme-value">{theme}</span>
      <button onClick={toggleTheme}>切换</button>
    </div>
  );
}

function renderTheme() {
  return render(
    <ThemeProvider>
      <ShowTheme />
    </ThemeProvider>
  );
}

describe('ThemeContext — 暗色模式切换', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.removeAttribute('data-theme');
  });

  afterEach(() => {
    cleanup();
  });

  it('默认主题为 light', () => {
    renderTheme();
    expect(screen.getByTestId('theme-value')).toHaveTextContent('light');
  });

  it('初始 data-theme 属性为 light', () => {
    renderTheme();
    expect(document.documentElement.getAttribute('data-theme')).toBe('light');
  });

  it('点击切换按钮变为 dark', async () => {
    renderTheme();
    await userEvent.click(screen.getByText('切换'));
    expect(screen.getByTestId('theme-value')).toHaveTextContent('dark');
  });

  it('切换后 data-theme 属性变为 dark', async () => {
    renderTheme();
    await userEvent.click(screen.getByText('切换'));
    expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
  });

  it('再次点击切回 light', async () => {
    renderTheme();
    await userEvent.click(screen.getByText('切换'));
    await userEvent.click(screen.getByText('切换'));
    expect(screen.getByTestId('theme-value')).toHaveTextContent('light');
  });

  it('主题持久化到 localStorage', async () => {
    renderTheme();
    await userEvent.click(screen.getByText('切换'));
    expect(localStorage.getItem('theme')).toBe('dark');
  });

  it('从 localStorage 恢复主题', () => {
    localStorage.setItem('theme', 'dark');
    renderTheme();
    expect(screen.getByTestId('theme-value')).toHaveTextContent('dark');
  });

  it('无效 localStorage 值回退到 light', () => {
    localStorage.setItem('theme', 'invalid');
    renderTheme();
    expect(screen.getByTestId('theme-value')).toHaveTextContent('light');
  });
});
