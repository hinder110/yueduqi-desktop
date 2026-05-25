// @vitest-environment jsdom
import { describe, it, expect, afterEach } from 'vitest';
import { render, screen, cleanup } from '@testing-library/react';
import StatusMessage from './StatusMessage';

describe('StatusMessage', () => {
  afterEach(cleanup);

  it('loading 为 true 时显示加载中', () => {
    render(<StatusMessage loading />);
    expect(screen.getByText('加载中...')).toBeDefined();
  });

  it('loading 为 true 且提供了 skeleton 时显示 skeleton', () => {
    render(
      <StatusMessage
        loading
        loadingSkeleton={<div data-testid="custom-skeleton">骨架屏</div>}
      />,
    );
    expect(screen.getByTestId('custom-skeleton')).toBeDefined();
    expect(screen.queryByText('加载中...')).toBeNull();
  });

  it('error 不为空时显示错误信息', () => {
    render(<StatusMessage error="网络错误，请重试" />);
    expect(screen.getByText('网络错误，请重试')).toBeDefined();
  });

  it('empty 为 true 时显示默认空文案', () => {
    render(<StatusMessage empty />);
    expect(screen.getByText('暂无数据')).toBeDefined();
  });

  it('empty 为 true 时显示自定义空文案', () => {
    render(<StatusMessage empty emptyText="没有找到任何内容" />);
    expect(screen.getByText('没有找到任何内容')).toBeDefined();
  });

  it('loading 优先级高于 error', () => {
    render(<StatusMessage loading error="出错了" />);
    expect(screen.getByText('加载中...')).toBeDefined();
    expect(screen.queryByText('出错了')).toBeNull();
  });

  it('error 优先级高于 empty', () => {
    render(<StatusMessage error="出错了" empty />);
    expect(screen.getByText('出错了')).toBeDefined();
    expect(screen.queryByText('暂无数据')).toBeNull();
  });

  it('所有状态都为 false 时渲染 null', () => {
    const { container } = render(<StatusMessage />);
    expect(container.innerHTML).toBe('');
  });

  it('同时设置 loading 和 error 时只显示 loading', () => {
    const { container } = render(<StatusMessage loading error="出错了" empty />);
    expect(screen.getByText('加载中...')).toBeDefined();
    expect(container.querySelector('.message.error')).toBeNull();
  });
});
