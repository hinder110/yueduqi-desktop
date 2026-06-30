import { SearchBooks, GetHotBooks, GetChapters, GetChapterContent } from '../wailsjs/go/main/App';

function ok<T>(data: T) { return { success: true as const, data }; }
function fail(error: string) { return { success: false as const, error }; }

// ── localStorage-based persistence ──
// Keeps bookshelf data and reading progress on the client side
// so the desktop app works offline without a backend.
const BOOKSHELF_KEY = 'bookshelf';

function readBookshelf(): any[] {
  try { const raw = localStorage.getItem(BOOKSHELF_KEY); return raw ? JSON.parse(raw) : []; }
  catch { return []; }
}

function writeBookshelf(items: any[]): void {
  localStorage.setItem(BOOKSHELF_KEY, JSON.stringify(items));
}

export async function fetchSearch(keyword: string, sourceKey?: string): Promise<any> {
  try { const data = await SearchBooks(keyword, sourceKey || 'guangyu'); return ok(data); }
  catch (e: any) { return fail(e?.message || '搜索失败'); }
}

export async function fetchHotBooks(): Promise<any> {
  try { const data = await GetHotBooks(); return ok(data); }
  catch (e: any) { return fail(e?.message || '获取热门推荐失败'); }
}

export async function fetchChapters(bookId: string, sourceKey?: string, innerSource?: string, innerTab?: string): Promise<any> {
  try {
    const data = await GetChapters(bookId, sourceKey || 'guangyu', innerSource || '', innerTab || '');
    return ok(data);
  }
  catch (e: any) { return fail(e?.message || '获取章节失败'); }
}

export async function fetchContent(bookId: string, itemId: string, sourceKey?: string, innerSource?: string, innerTab?: string): Promise<any> {
  try {
    const data = await GetChapterContent(bookId, itemId, sourceKey || 'guangyu', innerSource || '', innerTab || '');
    return ok(data);
  }
  catch (e: any) { return fail(e?.message || '获取正文失败'); }
}

export async function login(_u?: string, _p?: string): Promise<any> { return ok({ token: 'desktop', user: { id: 'local', username: 'local' } }); }
export async function register(_u?: string, _p?: string): Promise<any> { return ok({ id: 'local', username: 'local', created_at: new Date().toISOString() }); }
export async function addToBookshelf(book?: any): Promise<any> {
  try {
    const items = readBookshelf();
    // 防重复：同一 bookId 只保留一条
    const exists = items.find((item: any) => item.bookId === book?.bookId);
    if (exists) return ok(exists);
    const newItem = {
      ...book,
      id: Date.now(),
      addedAt: new Date().toISOString(),
      chapterIndex: 0,
      chapterItemId: '',
    };
    items.push(newItem);
    writeBookshelf(items);
    return ok(newItem);
  } catch (e: any) { return fail(e?.message || '加入书架失败'); }
}
export async function fetchBookshelf(): Promise<any> {
  try { return ok(readBookshelf()); }
  catch (e: any) { return fail(e?.message || '获取书架失败'); }
}
export async function removeFromBookshelf(id?: number): Promise<any> {
  try {
    const items = readBookshelf().filter((item: any) => item.id !== id);
    writeBookshelf(items);
    return ok({ message: '已移除' });
  } catch (e: any) { return fail(e?.message || '移除失败'); }
}
export async function updateProgress(id?: number, idx?: number, itemId?: string): Promise<any> {
  try {
    const items = readBookshelf();
    const item = items.find((item: any) => item.id === id);
    if (item) {
      item.chapterIndex = idx ?? item.chapterIndex;
      item.chapterItemId = itemId ?? item.chapterItemId;
      writeBookshelf(items);
    }
    return ok({ message: '已更新' });
  } catch (e: any) { return fail(e?.message || '更新进度失败'); }
}
