<template>
  <el-container class="app-container">
    <el-aside width="200px" class="sidebar">
      <div class="logo">
        <h2>OpenList-STRM</h2>
      </div>
      <el-menu
        :default-active="$route.path"
        router
        background-color="transparent"
        text-color="rgba(255,255,255,0.75)"
        active-text-color="#FFFFFF"
      >
        <el-menu-item index="/">
          <el-icon><Odometer /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        <el-menu-item index="/tasks">
          <el-icon><List /></el-icon>
          <span>任务管理</span>
        </el-menu-item>
        <el-menu-item index="/configs">
          <el-icon><Setting /></el-icon>
          <span>配置管理</span>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="header">
        <div class="header-title">{{ pageTitle }}</div>
        <div class="header-actions">
          <ThemeSwitcher />
        </div>
      </el-header>
      <el-main class="main-content">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import ThemeSwitcher from './components/ThemeSwitcher.vue'

const route = useRoute()

const pageTitle = computed(() => {
  const titles = {
    '/': '仪表盘',
    '/tasks': '任务管理',
    '/configs': '配置管理'
  }
  return titles[route.path] || 'OpenList-STRM'
})
</script>

<style scoped>
.app-container {
  height: 100vh;
  background: linear-gradient(135deg, var(--color-bg-start) 0%, var(--color-bg-end) 100%);
}

.sidebar {
  background: rgba(var(--color-sidebar-rgb), 0.95);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  border-right: 1px solid rgba(255, 255, 255, 0.2);
  box-shadow: 4px 0 24px rgba(var(--color-sidebar-rgb), 0.15);
}

.logo {
  height: 70px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #FFFFFF;
  border-bottom: 1px solid rgba(255, 255, 255, 0.15);
  background: rgba(255, 255, 255, 0.1);
}

.logo h2 {
  font-size: 17px;
  font-weight: 400;
  margin: 0;
  font-family: 'Varela Round', sans-serif;
  letter-spacing: 0.5px;
  text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.header {
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  box-shadow: 0 4px 16px rgba(var(--color-primary-rgb), 0.08);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 32px;
  border-bottom: 1px solid rgba(var(--color-primary-rgb), 0.1);
  border-radius: 0 0 24px 24px;
  margin: 0 16px;
}

.header-title {
  font-size: 22px;
  font-weight: 400;
  color: var(--color-text);
  font-family: 'Varela Round', sans-serif;
  letter-spacing: 0.3px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.main-content {
  background: transparent;
  padding: 32px;
  min-height: calc(100vh - 70px);
}

/* 菜单项样式 */
:deep(.el-menu-item) {
  border-radius: 12px;
  margin: 8px 12px;
  transition: all 0.2s ease;
}

:deep(.el-menu-item:hover) {
  background-color: rgba(255, 255, 255, 0.15) !important;
}

:deep(.el-menu-item.is-active) {
  background: rgba(255, 255, 255, 0.2) !important;
  font-weight: 600;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

:deep(.el-menu-item span) {
  font-weight: 400;
}
</style>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: 'Nunito Sans', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, 'PingFang SC', 'Microsoft YaHei', sans-serif;
  background: linear-gradient(135deg, var(--color-bg-start) 0%, var(--color-bg-end) 100%);
}

/* 标题使用 Varela Round，失败时回退到系统圆体 */
h1, h2, h3, h4, h5, h6, .heading {
  font-family: 'Varela Round', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, 'PingFang SC', 'Microsoft YaHei', sans-serif;
}

/* 代码/技术文本使用等宽字体 */
code, pre, .monospace {
  font-family: 'SF Mono', 'Monaco', 'Cascadia Code', 'Consolas', 'Courier New', monospace;
}
</style>
