import { SearchBooks, GetHotBooks, GetChapters, GetChapterContent } from '../wailsjs/go/main/App';

function ok<T>(data: T) { return { success: true as const, data }; }
function fail(error: string) { return { success: false as const, error }; }

export async function fetchSearch(keyword: string, sourceKey?: string): Promise<any> {
  try { const data = await SearchBooks(keyword, sourceKey || 'guangyu'); return ok(data); }
  catch (e: any) { return fail(e?.message || '搜索失败'); }
}

export async function fetchHotBooks(): Promise<any> {
  try { const data = await GetHotBooks(); return ok(data); }
  catch (e: any) { return fail(e?.message || '获取热门推荐失败'); }
}

export async function fetchChapters(bookId: string, _sourceKey?: string, _is?: string, _it?: string): Promise<any> {
  console.log('[fetchChapters] bookId:', bookId);
  try {
    const data = await GetChapters(bookId);
    console.log('[fetchChapters] result:', JSON.stringify(data).slice(0, 200));
    return ok(data);
  }
  catch (e: any) {
    console.error('[fetchChapters] error:', e?.message || e);
    return fail(e?.message || '获取章节失败');
  }
}

export async function fetchContent(bookId: string, itemId: string, _sourceKey?: string, _is?: string, _it?: string): Promise<any> {
  console.log('[fetchContent] bookId:', bookId, 'itemId:', itemId);
  try {
    const data = await GetChapterContent(bookId, itemId);
    console.log('[fetchContent] result:', JSON.stringify(data).slice(0, 200));
    return ok(data);
  }
  catch (e: any) {
    console.error('[fetchContent] error:', e?.message || e);
    return fail(e?.message || '获取正文失败');
  }
}

export async function login(_u?: string, _p?: string): Promise<any> { return ok({ token: 'desktop', user: { id: 'local', username: 'local' } }); }
export async function register(_u?: string, _p?: string): Promise<any> { return ok({ id: 'local', username: 'local', created_at: new Date().toISOString() }); }
export async function addToBookshelf(_book?: any): Promise<any> { return ok({ message: '已加入书架' }); }
export async function fetchBookshelf(): Promise<any> { return ok([]); }
export async function removeFromBookshelf(_id?: number): Promise<any> { return ok({ message: '已移除' }); }
export async function updateProgress(_id?: number, _idx?: number, _itemId?: string): Promise<any> { return ok({ message: '已更新' }); }
