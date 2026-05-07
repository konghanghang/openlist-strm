<template>
  <div class="dashboard">
    <el-row :gutter="20">
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card" @click="$router.push('/configs')">
          <div class="stat-content">
            <el-icon class="stat-icon" :style="{ color: 'var(--color-primary)' }"><Setting /></el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ stats.configCount }}</div>
              <div class="stat-label">配置数量</div>
              <div class="stat-extra">&nbsp;</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card" @click="$router.push('/tasks')">
          <div class="stat-content">
            <el-icon class="stat-icon" :style="{ color: 'var(--color-cta)' }"><List /></el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ stats.taskCount }}</div>
              <div class="stat-label">总任务数</div>
              <div class="stat-extra">&nbsp;</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card stat-card--static">
          <div class="stat-content">
            <el-icon class="stat-icon" :style="{ color: 'var(--color-secondary)' }"><InfoFilled /></el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ systemInfo.version }}</div>
              <div class="stat-label">系统版本</div>
              <div class="stat-extra">&nbsp;</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card stat-card--static">
          <div class="stat-content">
            <el-icon class="stat-icon" :style="{ color: 'var(--color-primary)' }"><Odometer /></el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ stats.uptime }}</div>
              <div class="stat-label">运行时间</div>
              <div class="stat-extra">启动于 {{ systemInfo.startTime }}</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="24">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>最近任务</span>
              <el-button type="text" @click="$router.push('/tasks')">查看全部</el-button>
            </div>
          </template>
          <el-table :data="recentTasks" style="width: 100%">
            <el-table-column prop="config_name" label="配置名称" min-width="160" show-overflow-tooltip />
            <el-table-column prop="mode" label="模式" width="90">
              <template #default="scope">
                <el-tag size="small" :type="scope.row.mode === 'full' ? 'warning' : ''">
                  {{ getModeText(scope.row.mode) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template #default="scope">
                <el-tag
                  :type="getStatusType(scope.row.status)"
                  size="small"
                >
                  {{ getStatusText(scope.row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="文件统计" min-width="200">
              <template #default="scope">
                <div class="file-stats">
                  <span class="stat-item">
                    <el-icon color="#67C23A"><CirclePlus /></el-icon>
                    {{ scope.row.files_created }}
                  </span>
                  <span class="stat-item">
                    <el-icon color="#F56C6C"><CircleClose /></el-icon>
                    {{ scope.row.files_deleted }}
                  </span>
                  <span class="stat-item">
                    <el-icon color="#E6A23C"><Remove /></el-icon>
                    {{ scope.row.files_skipped }}
                  </span>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="started_at" label="开始时间" width="170">
              <template #default="scope">
                {{ formatTime(scope.row.started_at) }}
              </template>
            </el-table-column>
            <el-table-column label="耗时" width="90">
              <template #default="scope">
                {{ calculateDuration(scope.row.started_at, scope.row.completed_at) }}
              </template>
            </el-table-column>

            <template #empty>
              <el-empty description="暂无任务记录" />
            </template>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../api'

const stats = ref({
  uptime: '0s',
  taskCount: 0,
  configCount: 0
})

const systemInfo = ref({
  version: '1.0.0',
  startTime: '-'
})

const recentTasks = ref([])

// Store uptime in seconds for auto-increment
let uptimeSeconds = 0
let uptimeInterval = null

const loadData = async () => {
  try {
    // Load status
    const status = await api.getStatus()
    uptimeSeconds = status.uptime
    stats.value.uptime = formatUptime(uptimeSeconds)
    systemInfo.value.version = status.version
    systemInfo.value.startTime = formatTime(status.start_time)

    // Load tasks (latest 10 records)
    const tasksData = await api.listTasks(1, 10)
    recentTasks.value = tasksData.tasks || []
    stats.value.taskCount = tasksData.total || 0

    // Load configs
    const configsData = await api.getConfigs()
    stats.value.configCount = configsData.configs ? configsData.configs.length : 0
  } catch (error) {
    console.error('Failed to load data:', error)
    ElMessage.error(`加载数据失败：${error.message}`)
  }
}

const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

const formatUptime = (seconds) => {
  if (!seconds) return '0s'
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = Math.floor(seconds % 60)
  return `${hours}h ${minutes}m ${secs}s`
}

const getModeText = (mode) => {
  const modeMap = {
    'incremental': '增量',
    'full': '全量'
  }
  return modeMap[mode] || mode
}

const getStatusType = (status) => {
  const types = {
    'completed': 'success',
    'running': 'warning',
    'failed': 'danger'
  }
  return types[status] || 'info'
}

const getStatusText = (status) => {
  const texts = {
    'completed': '已完成',
    'running': '运行中',
    'failed': '失败'
  }
  return texts[status] || status
}

const calculateDuration = (start, end) => {
  if (!start) return '-'
  if (!end) return '进行中'

  const startTime = new Date(start).getTime()
  const endTime = new Date(end).getTime()
  const seconds = Math.floor((endTime - startTime) / 1000)

  if (seconds < 60) return `${seconds}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`

  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  return `${hours}h ${minutes}m`
}

onMounted(() => {
  loadData()

  // 5 秒一次足够：人眼分辨不出更短的间隔，省下大量空跑
  uptimeInterval = setInterval(() => {
    uptimeSeconds += 5
    stats.value.uptime = formatUptime(uptimeSeconds)
  }, 5000)
})

onUnmounted(() => {
  if (uptimeInterval) {
    clearInterval(uptimeInterval)
  }
})
</script>

<style scoped>
.stat-card {
  cursor: pointer;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid rgba(255, 255, 255, 0.3);
  box-shadow: 0 8px 32px rgba(var(--color-primary-rgb), 0.12);
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  border-radius: 24px;
  overflow: hidden;
}

.stat-card:hover {
  transform: scale(1.02) translateY(-4px);
  box-shadow: 0 16px 48px rgba(var(--color-primary-rgb), 0.2);
  border-color: rgba(var(--color-primary-rgb), 0.4);
  background: rgba(255, 255, 255, 0.85);
}

/* 非可点击的 stat 卡：版本号、运行时间——只展示信息 */
.stat-card--static {
  cursor: default;
}

.stat-card--static:hover {
  transform: none;
  background: rgba(255, 255, 255, 0.7);
}

/* 列堆叠时（sm/xs 断点）补垂直间距，el-row gutter 只管水平 */
.el-col {
  margin-bottom: 16px;
}

@media (min-width: 992px) {
  .el-col {
    margin-bottom: 0;
  }
}

.stat-content {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 8px;
}

.stat-icon {
  font-size: 48px;
  opacity: 0.9;
  filter: drop-shadow(0 4px 8px rgba(var(--color-primary-rgb), 0.2));
}

.stat-info {
  flex: 1;
  min-height: 86px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.stat-value {
  font-size: 36px;
  font-weight: 400;
  color: var(--color-text);
  font-family: 'Varela Round', sans-serif;
  line-height: 1.1;
  background: linear-gradient(135deg, var(--color-primary) 0%, var(--color-secondary) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.stat-label {
  font-size: 14px;
  color: var(--color-text-muted);
  margin-top: 6px;
  font-weight: 600;
  letter-spacing: 0.3px;
}

.stat-extra {
  font-size: 11px;
  color: var(--color-text-muted);
  margin-top: 4px;
  min-height: 14px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 400;
  color: var(--color-text);
  font-family: 'Varela Round', sans-serif;
}

.file-stats {
  display: flex;
  gap: 20px;
}

.stat-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-family: 'Nunito Sans', sans-serif;
  font-size: 14px;
  font-weight: 600;
  padding: 4px 12px;
  background: rgba(var(--color-primary-rgb), 0.08);
  border-radius: 12px;
}

/* Dashboard 特有：表格行轻微缩放反馈 */
:deep(.el-table__row) {
  cursor: default;
}
</style>
