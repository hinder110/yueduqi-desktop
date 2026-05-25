interface StatusMessageProps {
  loading?: boolean;
  error?: string;
  empty?: boolean;
  emptyText?: string;
  loadingSkeleton?: React.ReactNode;
}

export default function StatusMessage({
  loading,
  error,
  empty,
  emptyText = '暂无数据',
  loadingSkeleton,
}: StatusMessageProps) {
  if (loading) {
    if (loadingSkeleton) return <>{loadingSkeleton}</>;
    return <div className="message loading">加载中...</div>;
  }
  if (error) return <div className="message error">{error}</div>;
  if (empty) return <div className="message empty">{emptyText}</div>;
  return null;
}
