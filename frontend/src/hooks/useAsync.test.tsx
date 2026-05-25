// @vitest-environment jsdom
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useAsync } from './useAsync';
import type { ApiResponse } from '../types';

describe('useAsync', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('初始状态 loading 为 false（非 immediate）', () => {
    const fn = vi.fn().mockResolvedValue({ success: true, data: 'hello' });
    const { result } = renderHook(() => useAsync(fn));

    expect(result.current.loading).toBe(false);
    expect(result.current.data).toBeNull();
    expect(result.current.error).toBe('');
  });

  it('immediate 模式初始 loading 为 true', () => {
    const fn = vi.fn().mockResolvedValue({ success: true, data: 'hello' });
    const { result } = renderHook(() => useAsync(fn, { immediate: true }));

    expect(result.current.loading).toBe(true);
  });

  it('immediate 模式自动执行并获取数据', async () => {
    const fn = vi.fn().mockResolvedValue({ success: true, data: ['a', 'b'] });
    const { result } = renderHook(() => useAsync<string[]>(fn, { immediate: true }));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(['a', 'b']);
    expect(fn).toHaveBeenCalled();
  });

  it('execute 手动调用并获取数据', async () => {
    const fn = vi.fn().mockResolvedValue({ success: true, data: 42 });
    const { result } = renderHook(() => useAsync<number>(fn));

    await act(async () => {
      await result.current.execute();
    });

    expect(result.current.data).toBe(42);
    expect(result.current.loading).toBe(false);
  });

  it('execute 可覆盖初始函数', async () => {
    const initialFn = vi.fn().mockResolvedValue({ success: true, data: 'initial' });
    const overrideFn = vi.fn().mockResolvedValue({ success: true, data: 'override' });
    const { result } = renderHook(() => useAsync<string>(initialFn));

    await act(async () => {
      await result.current.execute(overrideFn);
    });

    expect(result.current.data).toBe('override');
    expect(initialFn).not.toHaveBeenCalled();
    expect(overrideFn).toHaveBeenCalled();
  });

  it('API 返回 success:false 时设置 error', async () => {
    const fn = vi.fn().mockResolvedValue({ success: false, error: '服务器错误' });
    const { result } = renderHook(() => useAsync<string>(fn));

    await act(async () => {
      await result.current.execute();
    });

    expect(result.current.error).toBe('服务器错误');
    expect(result.current.data).toBeNull();
  });

  it('API 抛出异常时设置默认 error', async () => {
    const fn = vi.fn().mockRejectedValue(new Error('网络错误'));
    const { result } = renderHook(() => useAsync<string>(fn));

    await act(async () => {
      await result.current.execute();
    });

    expect(result.current.error).toBe('请求异常，请稍后重试');
  });

  it('setData 直接更新 data', () => {
    const fn = vi.fn();
    const { result } = renderHook(() => useAsync<string[]>(fn));

    act(() => {
      result.current.setData(['x', 'y']);
    });

    expect(result.current.data).toEqual(['x', 'y']);
  });

  it('setError 直接更新 error', () => {
    const fn = vi.fn();
    const { result } = renderHook(() => useAsync<string>(fn));

    act(() => {
      result.current.setError('自定义错误');
    });

    expect(result.current.error).toBe('自定义错误');
  });

  it('执行期间 loading 为 true', async () => {
    // 使用一个永不 resolve 的 promise 来测试 loading 状态
    let resolveFn!: (value: ApiResponse<string>) => void;
    const fn = vi.fn().mockImplementation(() => new Promise<ApiResponse<string>>((resolve) => {
      resolveFn = resolve;
    }));

    const { result } = renderHook(() => useAsync<string>(fn));

    act(() => {
      result.current.execute();
    });

    expect(result.current.loading).toBe(true);

    // 清理
    await act(async () => {
      resolveFn({ success: true, data: 'done' });
    });
  });

  it('data 为 undefined 时不更新 data（保持 null）', async () => {
    const fn = vi.fn().mockResolvedValue({ success: true, data: undefined });
    const { result } = renderHook(() => useAsync<string>(fn));

    await act(async () => {
      await result.current.execute();
    });

    // data 为 undefined 时不会被 set（保持 null）
    expect(result.current.data).toBeNull();
  });
});
