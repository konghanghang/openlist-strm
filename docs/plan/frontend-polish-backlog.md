# Plan: 前端打磨 Backlog

## Context

ui-ux-pro-max 探索（2026-05-07）后留下一组「改一处全局观感提升」的工程级改进点，
本次未全部落地——这份 backlog 把剩余事项按优先级排好，每条带定位、改法、预期。

设计规范（[`docs/design-system/index.md`](../design-system/index.md)）描述的是「项目当前应该是什么样」，
本 backlog 描述的是「项目现状与规范之间的差距」。每完成一条，相关代码与规范同时收口。

**来源：**

- ui-ux-pro-max skill 通用 UX 准则筛选（focus-visible、reduced-motion、对比度、skeleton、间距尺度）
- 项目当前代码与规范之间的差距比对（grep 验证至 commit `3f0aa8d`）

**非目标：**

- 不重做风格（Glassmorphism + 4 套主题保留）
- 不引新依赖（el-skeleton / el-empty 都是 element-plus 内置）
- 不堆解释性文案

## P0：改一处全局观感提升

### P0-1 stat 卡片视觉权重错配

**现状：** `web/src/views/Dashboard.vue:4-51`。4 张 stat 卡分两类：
- 可点击：「配置数量」「总任务数」（已加 click 跳转）
- 只读：「系统版本」「运行时间」（已加 `stat-card--static` 关闭 hover scale）

但视觉上四张卡完全一样——同样的渐变文字、同样的 icon 透明度。`stat-card--static` 只取消了交互动画，视觉权重未拉开。

**改法（`Dashboard.vue` scoped 样式）：**

```css
/* 不可点击的卡：去掉渐变文字 + icon 降透明 */
.stat-card--static .stat-value {
  background: none;
  -webkit-text-fill-color: var(--color-text);
  color: var(--color-text);
}

.stat-card--static .stat-icon {
  opacity: 0.5;
}

/* 可点击卡片：右下角加箭头微动效 */
.stat-card:not(.stat-card--static) .stat-info {
  position: relative;
}

.stat-card:not(.stat-card--static) .stat-info::after {
  content: '→';
  position: absolute;
  right: 0;
  bottom: 0;
  font-family: 'Varela Round', sans-serif;
  color: var(--color-primary);
  opacity: 0;
  transform: translateX(-4px);
  transition: all 0.2s ease;
}

.stat-card:not(.stat-card--static):hover .stat-info::after {
  opacity: 0.6;
  transform: translateX(0);
}
```

**预期：** 一眼分清「能点的」和「只看的」，hover 时可点击卡片的箭头滑入。

---

### P0-2 缺 prefers-reduced-motion 兜底（a11y）

**现状：** `web/src/styles/element-plus-theme.css` 全表检索无 `prefers-reduced-motion`。
项目大量 `transform: scale/translateY` + 0.2s/0.25s transitions，对前庭疾病敏感用户不友好，
且违反 WCAG 2.3.3。

**改法：** 在 `element-plus-theme.css` 末尾追加：

```css
@media (prefers-reduced-motion: reduce) {
  *,
  *::before,
  *::after {
    animation-duration: 0.01ms !important;
    transition-duration: 0.01ms !important;
    transform: none !important;
  }
}
```

**预期：** 系统级动画偏好生效，不影响其他用户。规范文档里已经写了这条，代码补齐即可。

---

### P0-3 加载状态太朴素（首屏 0 → 真实值跳变）

**现状：** `web/src/views/Dashboard.vue:139-160` 的 `loadData`。stat-value 初始化为 `'0s'`、`0`、`'1.0.0'`，
真实数据到达前用户会看到这些「假值」。

**改法：** 三处 stat-value 在 loading 状态下用 `el-skeleton-item` 占位：

```vue
<!-- 模板里 -->
<div class="stat-value">
  <el-skeleton-item v-if="loading" variant="text" style="width: 60%; height: 36px" />
  <span v-else>{{ stats.configCount }}</span>
</div>
```

`script setup` 里加 `const loading = ref(true)`，`loadData` 完成后置 `false`（包括 catch）。

**预期：** 消除「首屏 0 → 真实值」的误导跳变，与 Glassmorphism 风格协调。

---

## P1：明确性能与交互

### P1-4 表格行 hover 反馈太轻

**现状：** `web/src/styles/element-plus-theme.css:50-52`。行 hover 只有 `background-color: rgba(primary, 0.05)`，
亮度变化极弱，扫表格时锚点不强。

**改法：** 在 `element-plus-theme.css` 行 hover 规则上叠加左侧 indicator：

```css
.el-table__row:hover > td.el-table__cell:first-child {
  box-shadow: inset 3px 0 0 var(--color-primary);
}
```

**预期：** 当前 hover 行有清晰的左侧 3px 主题色条，表格扫读体验提升。

---

### P1-5 cron 预览没有 loading 反馈

**现状：** `web/src/views/Configs.vue:fetchCronPreview`。300ms 防抖窗口里 `nextRunTimes` 保持上一次值或为空，
用户切换控件后会有「无变化」的视觉空窗。

**改法：**

1. 加 `const previewLoading = ref(false)` ref
2. `fetchCronPreview` 进入时 `previewLoading.value = true`，结束置 `false`
3. 模板里 cron-preview 区域加：

```vue
<div v-if="previewLoading" class="cron-preview-label">
  <el-text type="info">计算中…</el-text>
</div>
<div v-else-if="previewError" class="cron-preview-error">...</div>
<div v-else-if="nextRunTimes.length > 0">...</div>
```

**预期：** 控件切换时有明确「计算中」提示，消除空窗。

---

### P1-6 focus-visible 状态没显式定义

**现状：** 项目大量改写了 button hover/active，但 `:focus-visible` 状态没显式覆盖，
键盘 Tab 用户可能看不到焦点位置。element-plus 默认的 outline 可能被改写过的 transform 干扰。

**改法：** 在 `element-plus-theme.css` 加：

```css
.el-button:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

.el-input__wrapper:focus-within {
  box-shadow: 0 0 0 3px rgba(var(--color-primary-rgb), 0.25) inset !important;
}

.el-menu-item:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: -2px;
}
```

**预期：** 键盘可达性回归，符合规范文档第 8 节要求。

---

## P2：细节收敛

### P2-7 el-empty 描述文案违反硬规则

**现状：** `web/src/views/Configs.vue:117`：

```vue
<el-empty description="暂无配置，请点击右上角「新增配置」添加路径映射" />
```

CLAUDE.md 项目硬规则禁止「页面说明文字」「设计者视角说明」。设计规范第 6.8 节也明确要求 ≤8 字。

**改法：**

```vue
<el-empty description="暂无配置" />
```

Tasks.vue 的「暂无任务记录」已合规，保持不变。

---

### P2-8 Dashboard stat-extra 用 `&nbsp;` 占位对齐

**现状：** `web/src/views/Dashboard.vue:11, 23, 35` 三处 `<div class="stat-extra">&nbsp;</div>`。
`.stat-info { min-height: 86px }` 已经在 scoped 里设了，占位 div 是冗余。

**改法：** 三处 `&nbsp;` 占位 div 直接删除。`stat-extra` 仅在「运行时间」卡保留实际文本。

**预期：** DOM 干净，少 4 个无意义节点。

---

### P2-9 后端错误文案中英文混搭

**现状：** axios 拦截器透出后端 `error` 字段，前端 catch 拼成「xxx 失败：${error.message}」。
但后端 `handlers.go` 的部分错误是英文，比如：

- `"mode must be 'incremental' or 'full'"`
- `"strm_mode must be 'alist_path' or 'http_url'"`
- `"failed to create mapping"`
- `"invalid cron expression: %v"`

最终用户看到的是「创建配置失败：mode must be 'incremental' or 'full'」这种中英混搭。

**改法（后端）：** 在 `backend/internal/api/handlers.go` 把面向用户的错误信息中文化：

```go
c.JSON(http.StatusBadRequest, gin.H{"error": "更新模式必须是「增量」或「全量」"})
c.JSON(http.StatusBadRequest, gin.H{"error": "STRM 模式必须是「Alist 路径」或「直链 URL」"})
c.JSON(http.StatusInternalServerError, gin.H{"error": "创建配置失败"})
c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Cron 表达式无效：%v", err)})
```

**预期：** 用户看到的错误从「创建配置失败：mode must be...」变成「创建配置失败：更新模式必须是「增量」或「全量」」，可读性提升。

注：这条涉及后端，单独立项 + 单独 commit 比较合适。

---

## P3：可选

### P3-10 Configs 操作列「更多」按钮 hover 反馈

**现状：** `.action-more` hover 时背景从 `rgba(primary, 0.06)` → `0.12`，但全局 `.el-table .el-button:hover` 已经叠加 `scale(1.04) + brightness(1.08)`，实际反馈足够。这一条可能已经够用，跑实测后再决定。

**判定：** 跑过实测发现明显不够再加。先记录，不立即动。

---

### P3-11 侧边栏菜单分组

**现状：** 3 个菜单平铺。如果未来增加到 ≥5 个菜单（如「定时任务」「日志」「系统设置」），
应该用 `el-menu-item-group` 拆分。

**判定：** 当前 3 个不需要。等菜单数增长后再做。

---

## 推荐执行顺序

按「改动量小但收益高」的顺序：

1. **P0-2**（reduced-motion，5 行 CSS，零风险）
2. **P2-7**（一行文案修改，零风险）
3. **P2-8**（删 3 行占位 div，零风险）
4. **P1-6**（focus-visible，~15 行 CSS）
5. **P1-4**（表格 hover indicator，3 行 CSS）
6. **P0-1**（stat 卡视觉分层，~25 行 CSS）
7. **P1-5**（cron 预览 loading，~10 行 JS + 模板）
8. **P0-3**（el-skeleton，模板 + script 改动较多）
9. **P2-9**（后端错误中文化，单独 PR）

前 5 条可一次打包提交（commit message 类似 `fix(web): 收口 a11y 与文案细节`）。
P0-1、P0-3 各自独立 commit。P2-9 单独后端 commit。

## 验证

每条都要：

1. `npm run build` 通过
2. 浏览器实测对应场景：
   - P0-1：打开 Dashboard，对比可点击卡与只读卡的视觉差异；hover 可点击卡看箭头
   - P0-2：浏览器 DevTools → Rendering → Emulate `prefers-reduced-motion: reduce`，验证动画停止
   - P0-3：刷新 Dashboard，看是否有「0 → 真实值」跳变
   - P1-4：扫表格观察 hover 行的左侧 indicator
   - P1-5：编辑配置，连续切换 cron 控件，观察「计算中…」提示
   - P1-6：键盘 Tab 遍历页面，焦点可见
   - P2-7：进入空状态的 Configs 页面（删完所有配置）观察文案
   - P2-8：DOM 检查 stat-extra 节点数

后端改动（P2-9）需要：

- `go test ./...` 通过
- 手动触发错误（提交无效 mode/strm_mode/cron），观察前端 message

## 设计规范同步

每条落地后，回写到 `docs/design-system/index.md`：

- P0-2 → 第 7.2 节已经写了规则，仅代码补齐，文档不动
- P0-1 → 第 9 节工程约束补一条「可点击卡片需有视觉差异」（可选）
- P1-6 → 第 8.1 节已写规则，代码补齐
- P2-7 → 第 6.8 节已写约束，代码修复后即合规
- P2-9 → 第 6.9 节加一句「后端错误信息必须为中文」

## 来源溯源

本 backlog 由 ui-ux-pro-max skill 通用 UX 准则筛选 + 项目代码现状对比生成。
skill 推荐的 OLED Dark + Fira Code 风格已在 OLED 实验中验证不合适，已回退；
本 backlog 仅保留与现有 Glassmorphism 浅色风格兼容的改进点。
