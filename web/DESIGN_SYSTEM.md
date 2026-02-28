# OpenList-STRM 设计系统

## 概述

本设计系统基于 UI/UX Pro Max 专业分析生成，采用 **Bento Grid + Glassmorphism** 混合风格，支持多主题切换。

## 设计风格

**风格名称**: Bento Grid + Glassmorphism（模块化网格 + 毛玻璃）

**关键特征**:
- Apple 风格的大圆角卡片（20-24px）
- 毛玻璃半透明效果（backdrop-blur）
- 模块化网格布局
- 明亮渐变背景
- 悬停时卡片微微放大（scale 1.02）

## 多主题系统

项目支持 4 套主题，通过 CSS 变量动态切换，用户选择保存在 localStorage。

### 天空蓝（默认）

| 角色 | 颜色值 | 说明 |
|------|--------|------|
| Primary | `#0EA5E9` | 天空蓝 |
| Secondary | `#38BDF8` | 浅天蓝 |
| CTA | `#F97316` | 暖橙 |
| Background | `#F0F9FF` → `#E0F2FE` | 天蓝渐变 |
| Text | `#0C4A6E` | 深蓝 |

### 青绿

| 角色 | 颜色值 | 说明 |
|------|--------|------|
| Primary | `#0D9488` | 青绿 |
| Secondary | `#14B8A6` | 浅青绿 |
| CTA | `#F97316` | 暖橙 |
| Background | `#F0FDFA` → `#CCFBF1` | 薄荷渐变 |
| Text | `#134E4A` | 深青绿 |

### 紫罗兰

| 角色 | 颜色值 | 说明 |
|------|--------|------|
| Primary | `#7C3AED` | 紫罗兰 |
| Secondary | `#A78BFA` | 浅紫 |
| CTA | `#CA8A04` | 金色 |
| Background | `#FAF5FF` → `#F3E8FF` | 淡紫渐变 |
| Text | `#4C1D95` | 深紫 |

### 玫瑰红

| 角色 | 颜色值 | 说明 |
|------|--------|------|
| Primary | `#E11D48` | 玫瑰红 |
| Secondary | `#FB7185` | 浅玫瑰 |
| CTA | `#2563EB` | 蓝色 |
| Background | `#FFF1F2` → `#FFE4E6` | 淡粉渐变 |
| Text | `#881337` | 深玫瑰 |

## CSS 变量

所有颜色通过 CSS 变量实现动态主题：

```css
--color-primary        /* 主色 */
--color-primary-rgb    /* 主色 RGB 值，用于 rgba() */
--color-secondary      /* 辅助色 */
--color-cta            /* 行动色 */
--color-cta-rgb        /* 行动色 RGB 值 */
--color-bg-start       /* 背景渐变起始色 */
--color-bg-end         /* 背景渐变结束色 */
--color-text           /* 主要文本色 */
--color-text-secondary /* 次要文本色 */
--color-text-muted     /* 辅助文本色 */
--color-sidebar-rgb    /* 侧边栏 RGB 值 */
```

## 字体系统

通过 `@fontsource` 自托管，不依赖外部 CDN：

```js
// main.js
import '@fontsource/nunito-sans/latin-400.css'
import '@fontsource/nunito-sans/latin-600.css'
import '@fontsource/nunito-sans/latin-700.css'
import '@fontsource/varela-round/latin-400.css'
```

- **标题**: Varela Round，fallback: 系统字体 + PingFang SC / Microsoft YaHei
- **正文**: Nunito Sans (400/600/700)，fallback: 同上
- **代码**: SF Mono / Monaco / Cascadia Code / Consolas

## 组件样式

### 卡片 (Glassmorphism)

```css
background: rgba(255, 255, 255, 0.7);
backdrop-filter: blur(10px);
border: 1px solid rgba(255, 255, 255, 0.3);
border-radius: 24px;
box-shadow: 0 8px 32px rgba(var(--color-primary-rgb), 0.12);
```

### 按钮

```css
background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%);
border-radius: 12px;
box-shadow: 0 4px 12px rgba(var(--color-primary-rgb), 0.3);
```

### 侧边栏

```css
background: rgba(var(--color-sidebar-rgb), 0.95);
backdrop-filter: blur(10px);
```

## 圆角系统

| 元素 | 圆角值 |
|------|--------|
| 卡片/对话框 | 24px |
| 按钮/输入框/菜单项 | 12px |
| 标签 | 8px |

## 交互效果

- 卡片悬停: `scale(1.02) translateY(-4px)`, 250ms
- 按钮悬停: `translateY(-2px)`, 200ms
- 表格行悬停: 背景色变化, 200ms

## 主题实现

- `src/theme.js` — 主题配置、CSS 变量注入、localStorage 持久化
- `src/components/ThemeSwitcher.vue` — 头部主题切换器
- 初始化时校验 localStorage 中的主题 key，不合法自动回退默认主题
