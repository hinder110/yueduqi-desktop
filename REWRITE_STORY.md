# 从 TypeScript 到 Go 桌面应用：难点与突破

## 背景

原来 yueduqi 只有一个 TypeScript 版本（Express + React），后来拆成了三个独立项目：

| 项目 | 技术 | 用途 |
|------|------|------|
| yueduqi | TypeScript | 原版，不动 |
| yueduqi-go | Go | 服务端，Docker 部署 |
| yueduqi-desktop | Go + Wails | 桌面应用，双击运行 |

---

## 难点 1：不是翻译，是重写

**问题：** 一开始以为把 TypeScript 代码一行行翻译成 Go 就行，其实不对。

TypeScript 的写法在 Go 里很别扭：
- TS 用 class、继承、装饰器 → Go 没 class，用 struct + interface
- TS 用 try/catch → Go 里 error 就是返回值
- TS 用 Promise.any 做并发 → Go 用 goroutine + channel

**突破：** 扔掉翻译思维。每段逻辑问自己："这个问题的 Go 答案是什么？" 而不是 "这段 TS 代码怎么用 Go 写？"

---

## 难点 2：JSON 字段名兼容

**问题：** 前端期望的字段是 `bookId`、`sourceKey`（camelCase），但 Go 默认导出大写字段。如果不小心漏了 json tag，前端立刻解析失败。

**突破：** Go 的 struct tag 一次性解决：
```go
type Book struct {
    BookID    string `json:"bookId"`    // Go 里叫 BookID，JSON 里叫 bookId
    SourceKey string `json:"sourceKey"` // 一一映射，编译时检查
}
```

检查清单：翻前端 `types.ts`，确保 Go 的 json tag 一个不漏。

---

## 难点 3：多镜像并发 + 取消

**问题：** 光遇 API 有 7 个镜像地址，TS 版用 `Promise.any` 同时发 7 个请求，谁快用谁。Go 里怎么做？

**突破：** goroutine + channel + context 取消：

```
同时起 7 个 goroutine 发请求 →
谁先成功就 cancel 其他 →
通过 channel 把结果传回主线程
```

Go 写这段比 TS 更清爽，因为 goroutine 就是干这个的。

---

## 难点 4：缓存从热变冷

**问题：** 切到 Go 版后第一次搜索觉得慢，以为是 Go 的问题。

**突破：** 测了个简单的对比：
- 冷缓存（第一次搜）：152ms（等 7 个外网 API）
- 热缓存（Redis 命中）：5ms

瓶颈在外网 API 的网络延迟，不在语言。Redis 重启后缓存空了而已，用几次就一样快。

---

## 难点 5：Wails 环境的坑

**问题：** 装 Wails 时遇到三个坑：

1. **Wails v3 vs v2**：v3 需要 webkit2gtk-6.0（系统没有），v2 需要 4.0（系统有 4.1）。最后用 v2 + 建 symlink 解决。
2. **CWD 陷阱**：Wails 命令必须在项目根目录跑，在子目录跑会报奇怪的错。
3. **frontend/wails.json 迷思**：vanilla 模板不需要这个文件，但用 react-ts 模板时 Wails 找这个文件。最后换 vanilla 模板重来一遍就正常了。

**突破：** 遇到模板问题别硬修，退回到已知能工作的 vanilla 模板，然后往上加东西。

---

## 难点 6：前端 API 层切换

**问题：** Web 版前端走 HTTP fetch 调后端。桌面版前端要直接调 Go 函数（Wails 绑定），不加 HTTP 层。

**突破：** 只改了一个文件 `api.ts`，其他页面代码不动：

```typescript
// Web 版：HTTP 请求
const res = await fetch('/api/search?keyword=斗罗');

// 桌面版：直接调 Go 函数
import { SearchBooks } from '../wailsjs/go/main/App';
const data = await SearchBooks('斗罗');
```

函数签名保持兼容（旧页面多传几个参数就忽略掉），所以 SearchPage、ChaptersPage 这些组件一行没改。

---

## 总结

最核心的教训两条：

1. **翻译是陷阱，思考才是正路。** 用目标语言的习惯写法，不要硬搬源语言的模式。
2. **先让最简版本跑通，再加复杂度。** Wails 模板坑多，就用 vanilla 先跑，再逐步加 React、parser、前端代码。

最终效果：同一套 React 前端，三个版本共享，Go 重写后内存小、启动快、编译成一个文件。
