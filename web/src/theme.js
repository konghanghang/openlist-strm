// 主题配置
export const themes = {
  skyBlue: {
    name: '天空蓝',
    id: 'skyBlue',
    colors: {
      primary: '#0EA5E9',
      primaryRgb: '14, 165, 233',
      secondary: '#38BDF8',
      cta: '#F97316',
      ctaRgb: '249, 115, 22',
      bgStart: '#F0F9FF',
      bgEnd: '#E0F2FE',
      text: '#0C4A6E',
      textSecondary: '#475569',
      textMuted: '#64748B',
      sidebarRgb: '14, 165, 233'
    }
  },
  teal: {
    name: '青绿',
    id: 'teal',
    colors: {
      primary: '#0D9488',
      primaryRgb: '13, 148, 136',
      secondary: '#14B8A6',
      cta: '#F97316',
      ctaRgb: '249, 115, 22',
      bgStart: '#F0FDFA',
      bgEnd: '#CCFBF1',
      text: '#134E4A',
      textSecondary: '#475569',
      textMuted: '#64748B',
      sidebarRgb: '13, 148, 136'
    }
  },
  purple: {
    name: '紫罗兰',
    id: 'purple',
    colors: {
      primary: '#7C3AED',
      primaryRgb: '124, 58, 237',
      secondary: '#A78BFA',
      cta: '#CA8A04',
      ctaRgb: '202, 138, 4',
      bgStart: '#FAF5FF',
      bgEnd: '#F3E8FF',
      text: '#4C1D95',
      textSecondary: '#475569',
      textMuted: '#64748B',
      sidebarRgb: '124, 58, 237'
    }
  },
  rose: {
    name: '玫瑰红',
    id: 'rose',
    colors: {
      primary: '#E11D48',
      primaryRgb: '225, 29, 72',
      secondary: '#FB7185',
      cta: '#2563EB',
      ctaRgb: '37, 99, 235',
      bgStart: '#FFF1F2',
      bgEnd: '#FFE4E6',
      text: '#881337',
      textSecondary: '#475569',
      textMuted: '#64748B',
      sidebarRgb: '225, 29, 72'
    }
  }
}

const DEFAULT_THEME = 'skyBlue'

// 严格校验主题 key
function isValidTheme(key) {
  return typeof key === 'string' && Object.prototype.hasOwnProperty.call(themes, key)
}

// 应用主题到 DOM
export function applyTheme(themeId) {
  if (!isValidTheme(themeId)) {
    themeId = DEFAULT_THEME
  }

  const theme = themes[themeId]
  const root = document.documentElement
  const colors = theme.colors

  // 设置 CSS 变量
  root.style.setProperty('--color-primary', colors.primary)
  root.style.setProperty('--color-primary-rgb', colors.primaryRgb)
  root.style.setProperty('--color-secondary', colors.secondary)
  root.style.setProperty('--color-cta', colors.cta)
  root.style.setProperty('--color-cta-rgb', colors.ctaRgb)
  root.style.setProperty('--color-bg-start', colors.bgStart)
  root.style.setProperty('--color-bg-end', colors.bgEnd)
  root.style.setProperty('--color-text', colors.text)
  root.style.setProperty('--color-text-secondary', colors.textSecondary)
  root.style.setProperty('--color-text-muted', colors.textMuted)
  root.style.setProperty('--color-sidebar-rgb', colors.sidebarRgb)

  try {
    localStorage.setItem('theme', themeId)
  } catch {
    // localStorage 不可用时静默忽略
  }
}

// 获取当前主题
export function getCurrentTheme() {
  try {
    const savedTheme = localStorage.getItem('theme')
    return isValidTheme(savedTheme) ? savedTheme : DEFAULT_THEME
  } catch {
    return DEFAULT_THEME
  }
}

// 初始化主题
export function initTheme() {
  const themeId = getCurrentTheme()
  applyTheme(themeId)
  return themeId
}
