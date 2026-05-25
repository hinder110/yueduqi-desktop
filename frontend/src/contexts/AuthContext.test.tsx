// @vitest-environment jsdom

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { render, screen, cleanup, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import '@testing-library/jest-dom/vitest';
import { AuthProvider, useAuth } from './AuthContext';
import { useContext } from 'react';

// localStorage mock
const store = new Map<string, string>();
vi.stubGlobal('localStorage', {
  getItem: vi.fn((key: string) => store.get(key) ?? null),
  setItem: vi.fn((key: string, value: string) => { store.set(key, value); }),
  removeItem: vi.fn((key: string) => { store.delete(key); }),
  clear: vi.fn(() => { store.clear(); }),
});

// Mock API 调用
vi.mock('../api', () => ({
  login: vi.fn(),
  register: vi.fn(),
}));

import { login as mockLogin, register as mockRegister } from '../api';

function ShowAuth() {
  const { user, loading, login, register, logout } = useAuth();
  return (
    <div>
      {loading && <span data-testid="loading">加载中...</span>}
      {user ? (
        <>
          <span data-testid="username">{user.username}</span>
          <button onClick={logout}>退出</button>
        </>
      ) : (
        <button
          onClick={async () => {
            const err = await login('test', '123456');
            if (err) document.querySelector('#error')!.textContent = err;
          }}
        >
          登录
        </button>
      )}
      <span id="error" data-testid="error" />
    </div>
  );
}

function renderAuth() {
  return render(
    <AuthProvider>
      <ShowAuth />
    </AuthProvider>
  );
}

describe('AuthContext — 认证状态管理', () => {
  beforeEach(() => {
    store.clear();
    vi.clearAllMocks();
  });

  afterEach(() => {
    cleanup();
  });

  // ========== 初始状态 ==========

  it('初始 loading 为 false，显示登录按钮', async () => {
    renderAuth();
    await waitFor(() => {
      expect(screen.queryByTestId('loading')).not.toBeInTheDocument();
    });
    expect(screen.getByText('登录')).toBeInTheDocument();
  });

  // ========== 登录 ==========

  it('登录成功后显示用户名', async () => {
    (mockLogin as ReturnType<typeof vi.fn>).mockResolvedValue({
      success: true,
      data: { token: 'fake-token', user: { id: '1', username: 'hinder' } },
    });
    renderAuth();
    await waitFor(() => {
      expect(screen.queryByTestId('loading')).not.toBeInTheDocument();
    });
    await userEvent.click(screen.getByText('登录'));
    await waitFor(() => {
      expect(screen.getByTestId('username')).toHaveTextContent('hinder');
    });
  });

  it('登录成功后 token 和 user 写入 localStorage', async () => {
    (mockLogin as ReturnType<typeof vi.fn>).mockResolvedValue({
      success: true,
      data: { token: 'fake-token', user: { id: '1', username: 'hinder' } },
    });
    renderAuth();
    await waitFor(() => {
      expect(screen.queryByTestId('loading')).not.toBeInTheDocument();
    });
    await userEvent.click(screen.getByText('登录'));
    await waitFor(() => {
      expect(store.get('token')).toBe('fake-token');
      expect(store.get('user')).toBe(JSON.stringify({ id: '1', username: 'hinder' }));
    });
  });

  it('登录失败显示错误', async () => {
    (mockLogin as ReturnType<typeof vi.fn>).mockResolvedValue({
      success: false,
      error: '密码错误',
    });
    renderAuth();
    await waitFor(() => {
      expect(screen.queryByTestId('loading')).not.toBeInTheDocument();
    });
    await userEvent.click(screen.getByText('登录'));
    await waitFor(() => {
      expect(screen.getByTestId('error')).toHaveTextContent('密码错误');
    });
  });

  it('登录失败不写入 localStorage', async () => {
    (mockLogin as ReturnType<typeof vi.fn>).mockResolvedValue({
      success: false,
      error: '密码错误',
    });
    renderAuth();
    await waitFor(() => {
      expect(screen.queryByTestId('loading')).not.toBeInTheDocument();
    });
    await userEvent.click(screen.getByText('登录'));
    await waitFor(() => {
      expect(screen.getByTestId('error')).toHaveTextContent('密码错误');
    });
    expect(localStorage.getItem('token')).toBeNull();
  });

  // ========== 退出 ==========

  it('退出后清除用户和 token', async () => {
    (mockLogin as ReturnType<typeof vi.fn>).mockResolvedValue({
      success: true,
      data: { token: 'fake-token', user: { id: '1', username: 'hinder' } },
    });
    renderAuth();
    await waitFor(() => {
      expect(screen.queryByTestId('loading')).not.toBeInTheDocument();
    });
    await userEvent.click(screen.getByText('登录'));
    await waitFor(() => {
      expect(screen.getByTestId('username')).toHaveTextContent('hinder');
    });
    await userEvent.click(screen.getByText('退出'));
    expect(screen.getByText('登录')).toBeInTheDocument();
    expect(localStorage.getItem('token')).toBeNull();
    expect(localStorage.getItem('user')).toBeNull();
  });

  // ========== localStorage 恢复会话 ==========

  it('从 localStorage 恢复已登录会话', () => {
    store.set('token', 'saved-token');
    store.set('user', JSON.stringify({ id: '2', username: 'olduser' }));
    renderAuth();
    // 因为 useEffect 是异步的，但 setLoading(false) 会同步触发重渲染
    waitFor(() => {
      expect(screen.getByTestId('username')).toHaveTextContent('olduser');
    });
  });

  it('localStorage 中无 token 则显示登录按钮', async () => {
    renderAuth();
    await waitFor(() => {
      expect(screen.getByText('登录')).toBeInTheDocument();
    });
  });

  it('localStorage 中 user 数据损坏则清除并显示登录按钮', async () => {
    store.set('token', 'bad-token');
    store.set('user', '{broken json');
    renderAuth();
    await waitFor(() => {
      expect(localStorage.getItem('token')).toBeNull();
      expect(localStorage.getItem('user')).toBeNull();
      expect(screen.getByText('登录')).toBeInTheDocument();
    });
  });

  // ========== useAuth 不在 Provider 内 ==========

  it('useAuth 在 Provider 外部使用抛出异常', () => {
    function BadComponent() {
      useAuth();
      return null;
    }
    expect(() => render(<BadComponent />)).toThrow('useAuth must be inside AuthProvider');
  });
});
