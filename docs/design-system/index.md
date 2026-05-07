# 前端设计规范

OpenList-STRM 前端的视觉与工程基线，覆盖色彩、字体、间距、组件、微交互、可访问性、反模式。
本文件是当前代码的事实快照——出现冲突时以 `web/src/` 实际实现为准，并同步回这里。

适用范围：

- Stack：Vue 3 + Vite + Element Plus + Vue Router
- 范围：`web/` 目录下所有页面、组件、样式
- 不在本规范覆盖：后端接口形态、CI/CD、部署

## 1. 设计语言

### 1.1 风格基调

**Bento Grid + Glassmorphism** 混合。Apple 风格大圆角卡片 + 半透明毛玻璃 + 模块化网格 + 明亮渐变背景。

关键特征：

- 卡片圆角 24px，半透明白底叠 `backdrop-filter: blur(10px)`
- 主背景使用主题色渐变（`bg-start` → `bg-end`）
- 4 套色彩主题 CSS 变量驱动，运行时切换
- 整体「轻、亮、有呼吸感」，避免厚重边框和粗实线

### 1.2 设计原则

1. **数据结构优先**：能用语义结构（标题、表格、标签）表达层级的，不靠装饰。
2. **少即是多**：面向用户的页面默认禁止堆解释性文案。能用标题、标签、按钮文案讲清楚的，不额外补一句"设计者视角说明"。
3. **一致性 > 局部巧思**：同一组件在不同页面应表现一致，特例必须有清楚的理由。
4. **CSS 变量唯一来源**：所有颜色经 `--color-*` 变量。任何写死的色值都视为缺陷。
5. **改一处全局生效**：跨页面的 element-plus 覆盖样式只能写在 `web/src/styles/element-plus-theme.css`，不在 scoped 块里重复。

## 2. 色彩系统

### 2.1 主题清单

4 套主题在 `web/src/theme.js` 定义，默认 `skyBlue`，用户切换后存 `localStorage`。

| 主题 ID | 名称 | Primary | Secondary | CTA | 背景渐变 | 主文本 |
|---------|------|---------|-----------|-----|----------|--------|
| `skyBlue` | 天空蓝 | `#0EA5E9` | `#38BDF8` | `#F97316` | `#F0F9FF` → `#E0F2FE` | `#0C4A6E` |
| `teal` | 青绿 | `#0D9488` | `#14B8A6` | `#F97316` | `#F0FDFA` → `#CCFBF1` | `#134E4A` |
| `purple` | 紫罗兰 | `#7C3AED` | `#A78BFA` | `#CA8A04` | `#FAF5FF` → `#F3E8FF` | `#4C1D95` |
| `rose` | 玫瑰红 | `#E11D48` | `#FB7185` | `#2563EB` | `#FFF1F2` → `#FFE4E6` | `#881337` |

### 2.2 CSS 变量

样式里只允许使用变量，不写死十六进制：

```css
--color-primary           /* 主色，用于按钮主操作、链接、强调文本 */
--color-primary-rgb       /* 主色 RGB 三元组，用于 rgba() 透明叠加 */
--color-secondary         /* 辅助色，渐变第二色 */
--color-cta               /* 行动色（与 primary 形成对比，用于"全部生成"等强动作）*/
--color-cta-rgb
--color-bg-start          /* 主背景渐变起点 */
--color-bg-end            /* 主背景渐变终点 */
--color-text              /* 主文本 */
--color-text-secondary    /* 次文本（标签、辅助说明）*/
--color-text-muted        /* 弱文本（提示、占位）*/
--color-sidebar-rgb       /* 侧边栏底色 RGB 三元组 */
```

### 2.3 使用约束

- 文本色用 `var(--color-text*)`，不用 `#909399`、`#475569` 这类硬编码值。历史代码里如还存在请就地清理。
- 透明叠加用 `rgba(var(--color-primary-rgb), 0.08)` 形式，不预先在变量里指定透明度。
- 状态色（success / warning / danger）element-plus 内置的 `#10B981` / `#F59E0B` / `#EF4444` 渐变保留，不主题化（否则破坏语义）。
- 状态色仅用于 tag、message、按钮的 success/warning/danger 变体，不泛化到主结构。

## 3. 字体系统

### 3.1 字族

通过 `@fontsource` 自托管，不依赖外部 CDN。`font-display: swap` 已在包内默认启用。

```js
// main.js
import '@fontsource/nunito-sans/400.css'
import '@fontsource/nunito-sans/600.css'
import '@fontsource/nunito-sans/700.css'
import '@fontsource/varela-round/400.css'
```

| 用途 | 字族 | Fallback |
|------|------|----------|
| 标题、按钮、对话框标题 | `Varela Round` | 系统圆体 / PingFang SC / Microsoft YaHei |
| 正文、表格 | `Nunito Sans` (400/600/700) | 系统字体栈 |
| 代码、路径、Cron 表达式、UUID | `SF Mono` / `Monaco` / `Cascadia Code` | `Consolas`、`Courier New` |

### 3.2 字号尺度

| 用途 | 尺寸 | 字重 | 字族 |
|------|------|------|------|
| 大数字（stat-value） | 36px | 400 | Varela Round |
| 卡片标题 / 对话框标题 | 18-22px | 400 | Varela Round |
| 正文、按钮 | 14px | 600 | Nunito Sans |
| 表头 | 13px | 700 | Nunito Sans，UPPERCASE，letter-spacing 0.5px |
| 标签、辅助说明 | 12-13px | 600 | Nunito Sans |
| 时间 / 路径等代码型文本 | 12-13px | 400 | SF Mono |

### 3.3 行高

- 正文：1.5
- 列表项 / 时间预览：1.8
- 大数字 / 标题：1.1

## 4. 间距、圆角、阴影

### 4.1 间距尺度

只用 4 / 8 / 12 / 16 / 20 / 24 / 32 七个值。其他数字视为缺陷。

| 场景 | 值 |
|------|----|
| 表格行垂直 padding | 16px |
| 表头垂直 padding | 14px |
| 卡片 body padding | 24px |
| 卡片 header padding | 20px 24px |
| 对话框 body padding | 24px |
| 对话框 footer padding | 16px 24px |
| 行内元素之间 | 8 / 12 / 16px |
| 栅格 gutter | 20px |
| 整页 main padding | 32px |

### 4.2 圆角尺度

| 元素 | 圆角 |
|------|------|
| 卡片、对话框 | 24px |
| 按钮、输入框、菜单项、input 类 | 12px |
| Alert、空状态 | 16px |
| 标签 | 8px |
| 头像、小色块 | 6px |
| 分页 active 项 | 10px |

### 4.3 阴影 / 模糊度

模糊度统一 `blur(10px)`（卡片、对话框、侧边栏、按钮）。

阴影分四档：

```css
/* sm —— 小元素：按钮 */
0 4px 12px rgba(var(--color-primary-rgb), 0.3)

/* md —— 卡片默认 */
0 8px 32px rgba(var(--color-primary-rgb), 0.12)

/* lg —— 卡片 hover、按钮 hover */
0 12px 48px rgba(var(--color-primary-rgb), 0.18)
0 8px 20px rgba(var(--color-primary-rgb), 0.4)

/* xl —— 对话框 */
0 24px 64px rgba(var(--color-primary-rgb), 0.2)
```

阴影颜色用 `--color-primary-rgb`，让阴影随主题色变化。

## 5. 响应式

### 5.1 断点

沿用 element-plus 默认断点：

| 名称 | 区间 | 典型设备 |
|------|------|----------|
| `xs` | < 768px | 手机 |
| `sm` | ≥ 768px | 平板竖 |
| `md` | ≥ 992px | 平板横 / 小屏笔记本 |
| `lg` | ≥ 1200px | 主流笔记本 |
| `xl` | ≥ 1920px | 大屏 |

### 5.2 应用规范

- **栅格**：使用 `<el-col :xs="24" :sm="12" :md="6">` 形式，`md` 起单行展示 4 卡。不用 `:span="6"` 写死。
- **栅格垂直间距**：`<el-row :gutter="20">` 只控制水平间距，列堆叠时需在 scoped 加 `@media (min-width: 992px) { .el-col { margin-bottom: 0 } }`。
- **表格列宽**：关键列用 `min-width` + `show-overflow-tooltip`，操作列固定 `width` 并加 `fixed="right"`。绝不能所有列写死 `width` 让表格只占 800px，右侧大片留白。
- **整页最小宽**：以 1366px 笔记本可读为底线。低于该宽度允许出现侧滑布局，但功能必须可达。

## 6. 组件规范

### 6.1 卡片（el-card）

公共样式在 `web/src/styles/element-plus-theme.css`，覆盖了 element-plus 的默认底色、阴影、圆角、padding。页面只在需要不同悬停效果时（如 Dashboard 的 stat-card 用 `scale(1.02) translateY(-4px)`）写 scoped 覆盖。

### 6.2 按钮（el-button）

| 变体 | 用途 | 视觉 |
|------|------|------|
| `primary` | 页面主操作 | 主题渐变 + 阴影 sm |
| `success` | 创建类、可操作但非首要 | 绿渐变 |
| `warning` | 编辑、修改类 | 橙渐变 |
| `danger` | 删除等不可逆 | 红渐变 |
| `text` | 列表内次要操作（如「查看详情」） | 主题色文本，hover 加浅底 |
| 默认 | 取消、刷新 | 浅底深字 |

按钮 hover：

- 表格**外**按钮：`translateY(-2px)` + 阴影加深
- 表格**内**按钮：`scale(1.04)` + `brightness(1.08)` + 阴影加深（`translateY` 会被 `.el-table .cell` 的 `overflow: hidden` 切掉）

按钮 active：

- 表格外：`translateY(0)`
- 表格内：`scale(1)`

### 6.3 表格（el-table）

强制规范：

- 关键列必须 `min-width` 且 `show-overflow-tooltip`，让 element-plus 自动分配剩余宽度。
- 状态、模式、固定标签类列用 `width="80~110"` 锁死。
- 操作列加 `fixed="right"`，宽度根据按钮数量定（单按钮 110、主+下拉 240）。
- 所有表格必须配 `<template #empty><el-empty description="..." /></template>`，不用默认 "No Data"。
- 行高度由公共样式保证（约 58px），不在页面内重复 padding 设置。

### 6.4 操作列分层

- 主操作（生成、查看详情）独立明示，用 primary / text 按钮。
- 次操作（编辑、通知、复制）折叠到「更多」`el-dropdown`。
- 危险操作（删除）放进同一个 dropdown 但用 `divided` 与 `class="action-delete"` 单独标红。

历史教训：早期 Configs 把「生成 / 通知 / 编辑 / 删除」四个全彩渐变按钮排排坐，主次不分、视觉权重一致，且 1366 屏会被挤爆。**禁止再退回这种形态**。

### 6.5 对话框（el-dialog）

- 圆角 24px，毛玻璃边框，header 浅底主题色叠加。
- footer 按钮顺序：取消（左，默认 button）→ 主操作（右，primary）。
- 表单 label-width 固定 120px，过长 label 不允许换行。

### 6.6 表单（el-form）

- 输入框圆角 12px，focus 时用 inset 阴影 `box-shadow: 0 0 0 2px rgba(primary, 0.4) inset`。
- 路径、Cron 表达式等"代码型"字段的 input 用 SF Mono。
- 字段提示文案颜色走 `--color-text-muted`，不写 `#909399`。

### 6.7 标签（el-tag）

- 圆角 8px，无 border，padding 4px 12px，font-weight 600。
- 用法语义：状态走 success/warning/danger/info，业务区分（如增量/全量）走 type 默认或 warning。

### 6.8 空状态

统一用 `el-empty`。`description` 文案保持简短（≤8 字），不要写「请点击右上角……添加」这类引导，违反"禁止解释性文案"。允许的例子："暂无配置" / "暂无任务记录"。

### 6.9 加载与反馈

| 场景 | 反馈方式 |
|------|----------|
| 表格列表加载 | `v-loading` |
| 异步操作按钮 | `:loading="..."` |
| 数据未到的统计数字（首屏） | `el-skeleton` |
| 操作成功 | `ElMessage.success(具体结果)` |
| 操作失败 | `ElMessage.error(\`xxx 失败：${error.message}\`)` |

错误信息必须把后端真实原因透出，不能吞掉。axios 响应拦截器（`web/src/api/index.js`）已经把后端 `{error: ...}` 标准化到 `error.message`，调用方直接读即可。

## 7. 微交互

### 7.1 transition 节奏

| 用途 | 时长 | 缓动 |
|------|------|------|
| 颜色、阴影变化 | 0.2s | ease |
| 大组件位移 / 缩放 | 0.25s | `cubic-bezier(0.4, 0, 0.2, 1)` |
| 菜单项 hover | 0.2s | ease |

短于 0.15s 用户察觉不到，长于 0.3s 显得拖沓。

### 7.2 prefers-reduced-motion

`element-plus-theme.css` 内必须包含全局兜底：

```css
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
    transform: none !important;
  }
}
```

不针对单组件加，全局一段就够。

### 7.3 cursor

- 可点击元素必须 `cursor: pointer`。
- 反过来：设了 `cursor: pointer` 必须有点击行为（包括 router 跳转）。Dashboard 的不可点击 stat 卡用 `.stat-card--static` 修饰类显式去除。

## 8. 可访问性

### 8.1 focus-visible

所有可交互元素必须有可见的键盘焦点态。`element-plus-theme.css` 强制覆盖：

```css
.el-button:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}
```

不要用 `outline: none` 而不补替代样式。

### 8.2 对比度

- 正文文本与背景对比 ≥ 4.5:1（WCAG AA）。
- 大号文本（≥18px / 14px bold）≥ 3:1。
- 浅色主题下禁止用 `--color-text-muted` 作为正文色。

### 8.3 键盘导航

- Tab 顺序与视觉顺序一致。
- 表单 Enter 提交，对话框 Esc 关闭——element-plus 默认行为，不要禁用。
- 图标按钮（仅 icon 无文字）必须有 `aria-label`。

## 9. 工程约束

### 9.1 公共样式 vs scoped 样式

**写在 `web/src/styles/element-plus-theme.css`：**

- 跨页共用的 element-plus 组件覆盖：`.el-card / .el-table / .el-button / .el-dialog / .el-form / .el-tag / .el-pagination / .el-alert`
- 必须用更具体的 specificity（如 `td.el-table__cell` 而非 `.el-table td`）压过 element-plus 自身样式
- 一律不带 `:deep`，因为本身就是全局

**写在页面 scoped 内：**

- 页面独有的 layout（如 Dashboard 的 `.stat-card`、Configs 的 `.action-cell`）
- 公共样式的「该页特例」（如 Tasks 表格行的 `cursor: pointer`）

### 9.2 何时用 :deep

- Vue scoped 样式无法穿透到 element-plus 内部 DOM，需要 `:deep(.el-xxx)`。
- 但若该 :deep 选择器**所有页都需要**，应该搬去公共样式文件，而不是在每页重复。
- Teleport 出去的 popper（如 `el-dropdown` 默认 teleport）`:deep` 不生效，必须用 `popper-class` + 同文件内非 scoped `<style>` 块。

### 9.3 主题持久化与回退

- `theme.js` 启动时 `initTheme()` 校验 localStorage，非法值回落到 `skyBlue`。
- 切换器使用 `popper-class` 隔离样式，不要写 `:teleported="false"`（会被 header 的 `backdrop-filter` 创建的 stacking context 吃掉）。

### 9.4 Element Plus 选择器优先级

element-plus 内部用 `.el-table .el-table__cell`（specificity 0,2,0），覆盖时如果只写 `.el-table td`（0,1,1）会被吃掉。要写更具体或合并双 class 的选择器（`td.el-table__cell` = 0,1,1，但更可靠的是 `.el-table td.el-table__cell` = 0,2,1）。

## 10. 反模式清单

下列做法在历史代码里出现过且已经修复，今后**禁止再写**。

### 视觉层

- ❌ 操作列堆 4 个全彩渐变按钮排排坐 → ✅ 主操作 + dropdown 更多
- ❌ stat 卡片 `cursor: pointer` 但没有 click handler → ✅ 二选一
- ❌ 4 张 stat 卡片用同等视觉权重，可点击与只读混在一起 → ✅ 只读卡用 `--static` 修饰类区分
- ❌ stat-extra 用 `&nbsp;` 占位对齐 → ✅ 父容器 `min-height`
- ❌ 表格列宽全部写死 `width` → ✅ 关键列 `min-width` + `show-overflow-tooltip`
- ❌ 空状态用 `el-alert` 叠在表格外 → ✅ 表格内 `#empty` slot 用 `el-empty`
- ❌ 空状态描述写「请点击右上角……」 → ✅ 简短到「暂无配置」

### 样式层

- ❌ 三个页面各自重写 200 行 `:deep()` 卡片 / 表格样式 → ✅ 统一搬去 `element-plus-theme.css`
- ❌ 硬编码 `color: #909399` / `#475569` / `#94A3B8` → ✅ `var(--color-text-muted)` / `var(--color-text-secondary)`
- ❌ `transform: translateY(-2px)` 加在表格内按钮上 → ✅ 表格内按钮用 `scale(1.04)`
- ❌ 用 `.el-table td` 覆盖 padding（输给 element-plus 自身的 `.el-table .el-table__cell`）→ ✅ 用 `td.el-table__cell`
- ❌ 用 `setInterval(1000)` 刷新 uptime → ✅ `setInterval(5000)` 累加 5

### 行为层

- ❌ catch 里写死「xxx 失败」吞掉后端 message → ✅ 拼 `${error.message}`
- ❌ 主题切换器 `:teleported="false"` 让下拉留在 backdrop-filter 容器内 → ✅ 默认 teleport，配 `popper-class`
- ❌ Cron 预览前端硬算（与后端 robfig/cron 实际触发时间会偏差）→ ✅ 调 `POST /api/cron/preview`
- ❌ 把"视频扩展名"写死在 UI 文案里限制用户 → ✅ "媒体扩展名" + `allow-create` 让用户自定义

### 文档层

- ❌ 在多个文档里维护同一份事实 → ✅ 单源 + 链接互引
- ❌ 文档名用大写（`DESIGN_SYSTEM.md`）→ ✅ 统一小写
- ❌ 写"以后再整理"占位段 → ✅ 直接删或改写

## 11. 引用与索引

代码事实：

- 主题：`web/src/theme.js`
- 公共样式：`web/src/styles/element-plus-theme.css`
- 主题切换：`web/src/components/ThemeSwitcher.vue`
- 入口：`web/src/main.js`
- 页面：`web/src/views/{Dashboard,Tasks,Configs}.vue`

相关规范文档：

- [系统架构](../system-architecture.md) —— 后端实现与前后端契约
- [开发指南](../development-guide.md) —— 阅读路径与最小验证
- [PRD](../PRD.md) —— 产品需求与功能边界

修改本文件时，必须同步修改相关代码或反过来——文档与代码不允许长期不一致。
