<template>
  <div class="dashboard">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <el-icon class="stat-icon" color="#E6A23C"><Setting /></el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ stats.configCount }}</div>
              <div class="stat-label">配置数量</div>
              <div class="stat-extra">&nbsp;</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <el-icon class="stat-icon" color="#67C23A"><List /></el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ stats.taskCount }}</div>
              <div class="stat-label">总任务数</div>
              <div class="stat-extra">&nbsp;</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <el-icon class="stat-icon" color="#909399"><InfoFilled /></el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ systemInfo.version }}</div>
              <div class="stat-label">系统版本</div>
              <div class="stat-extra">&nbsp;</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stat-card">
          <div class="stat-content">
            <el-icon class="stat-icon" color="#409EFF"><Odometer /></el-icon>
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
            <el-table-column prop="config_name" label="配置名称" width="150" />
            <el-table-column prop="mode" label="模式" width="100">
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
            <el-table-column label="文件统计" width="200">
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
            <el-table-column prop="started_at" label="开始时间" width="180">
              <template #default="scope">
                {{ formatTime(scope.row.started_at) }}
              </template>
            </el-table-column>
            <el-table-column label="耗时" width="100">
              <template #default="scope">
                {{ calculateDuration(scope.row.started_at, scope.row.completed_at) }}
              </template>
            </el-table-column>
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
    ElMessage.error('加载数据失败')
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

  // Start uptime auto-increment (every second)
  uptimeInterval = setInterval(() => {
    uptimeSeconds++
    stats.value.uptime = formatUptime(uptimeSeconds)
  }, 1000)
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
  transition: all 0.3s;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
}

.stat-content {
  display: flex;
  align-items: center;
}

.stat-icon {
  font-size: 48px;
  margin-right: 20px;
}

.stat-info {
  flex: 1;
}

.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-top: 4px;
}

.stat-extra {
  font-size: 12px;
  color: #C0C4CC;
  margin-top: 4px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.file-stats {
  display: flex;
  gap: 12px;
}

.stat-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.info-item {
  padding: 8px 0;
  display: flex;
  align-items: center;
}

.info-label {
  color: #909399;
  margin-right: 10px;
  min-width: 80px;
}

.info-value {
  color: #303133;
  font-weight: 500;
}
</style>
