<template>
  <el-dropdown
    trigger="click"
    popper-class="theme-switcher-popper"
    @command="handleThemeChange"
  >
    <el-button circle class="theme-button">
      <el-icon><Brush /></el-icon>
    </el-button>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item
          v-for="theme in themeList"
          :key="theme.id"
          :command="theme.id"
          :class="{ 'is-active': currentTheme === theme.id }"
        >
          <div class="theme-item">
            <div class="theme-color" :style="{ background: theme.colors.primary }"></div>
            <span>{{ theme.name }}</span>
            <el-icon v-if="currentTheme === theme.id" class="check-icon"><Check /></el-icon>
          </div>
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { themes, applyTheme, getCurrentTheme } from '../theme'
import { Brush, Check } from '@element-plus/icons-vue'

const currentTheme = ref('skyBlue')
const themeList = Object.values(themes)

const handleThemeChange = (themeId) => {
  applyTheme(themeId)
  currentTheme.value = themeId
}

onMounted(() => {
  currentTheme.value = getCurrentTheme()
})
</script>

<style scoped>
.theme-button {
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.3);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transition: all 0.2s ease;
}

.theme-button:hover {
  transform: scale(1.05);
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.15);
}

.theme-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 4px 0;
  min-width: 140px;
}

.theme-color {
  width: 20px;
  height: 20px;
  border-radius: 6px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.check-icon {
  margin-left: auto;
  color: var(--color-primary, #0EA5E9);
  font-weight: bold;
}
</style>

<!-- popper teleport 到 body 后 :deep 失效，active 样式必须走全局 -->
<style>
.theme-switcher-popper .el-dropdown-menu__item.is-active {
  background-color: rgba(var(--color-primary-rgb, 14, 165, 233), 0.1);
  color: var(--color-primary, #0EA5E9);
}
</style>
