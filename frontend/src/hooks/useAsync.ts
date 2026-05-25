import { useState, useCallback, useEffect, useRef } from 'react';
import type { ApiResponse } from '../types';

interface UseAsyncReturn<T> {
  data: T | null;
  loading: boolean;
  error: string;
  execute: (fn?: () => Promise<ApiResponse<T>>) => Promise<void>;
  setData: (data: T | null) => void;
  setError: (error: string) => void;
}

export function useAsync<T>(
  initialFn?: () => Promise<ApiResponse<T>>,
  options?: { immediate?: boolean },
): UseAsyncReturn<T> {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(!!options?.immediate);
  const [error, setError] = useState('');
  const mountedRef = useRef(true);

  useEffect(() => {
    mountedRef.current = true;
    return () => { mountedRef.current = false; };
  }, []);

  const execute = useCallback(async (fn?: () => Promise<ApiResponse<T>>) => {
    setLoading(true);
    setError('');
    try {
      const res = await (fn ?? initialFn)!();
      if (!mountedRef.current) return;
      if (res.success && res.data !== undefined) {
        setData(res.data);
      } else {
        setError(res.error ?? '加载失败');
      }
    } catch {
      if (mountedRef.current) setError('请求异常，请稍后重试');
    } finally {
      if (mountedRef.current) setLoading(false);
    }
  }, [initialFn]);

  useEffect(() => {
    if (options?.immediate && initialFn) {
      execute(initialFn);
    }
  }, []);

  return { data, loading, error, execute, setData, setError };
}
