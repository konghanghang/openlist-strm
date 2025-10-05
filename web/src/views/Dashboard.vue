<template>
  <div class="dashboard">
    <el-row :gutter="20">
      <el-col :span="8">
        <el-card class="stat-card">
          <div class="stat-content">
            <el-icon class="stat-icon" color="#409EFF"><Odometer /></el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ stats.uptime }}</div>
              <div class="stat-label">运行时间</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card class="stat-card">
          <div class="stat-content">
            <el-icon class="stat-icon" color="#67C23A"><List /></el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ stats.taskCount }}</div>
              <div class="stat-label">总任务数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card class="stat-card">
          <div class="stat-content">
            <el-icon class="stat-icon" color="#E6A23C"><Setting /></el-icon>
            <div class="stat-info">
              <div class="stat-value">{{ stats.configCount }}</div>
              <div class="stat-label">配置数量</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>快速操作</span>
            </div>
          </template>
          <el-button
            type="primary"
            @click="handleGenerate"
            :loading="generating"
            style="width: 100%; margin-bottom: 10px;"
          >
            <el-icon><VideoPlay /></el-icon>
            <span>生成所有 STRM 文件</span>
          </el-button>
          <el-alert
            v-if="generateResult"
            :title="generateResult.message"
            :type="generateResult.type"
            :closable="false"
            style="margin-top: 10px;"
          />
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>系统信息</span>
            </div>
          </template>
          <div class="info-item">
            <span class="info-label">版本：</span>
            <span class="info-value">{{ systemInfo.version }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">启动时间：</span>
            <span class="info-value">{{ systemInfo.startTime }}</span>
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
            <el-table-column prop="config_name" label="配置名称" width="180" />
            <el-table-column prop="mode" label="模式" width="100">
              <template #default="scope">
                <el-tag size="small">{{ scope.row.mode }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="120">
              <template #default="scope">
                <el-tag
                  :type="scope.row.status === 'completed' ? 'success' : scope.row.status === 'failed' ? 'danger' : 'warning'"
                  size="small"
                >
                  {{ scope.row.status }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="files_created" label="创建文件" width="100" />
            <el-table-column prop="started_at" label="开始时间">
              <template #default="scope">
                {{ formatTime(scope.row.started_at) }}
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
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
const generating = ref(false)
const generateResult = ref(null)

const loadData = async () => {
  try {
    // Load status
    const status = await api.getStatus()
    stats.value.uptime = formatUptime(status.uptime)
    systemInfo.value.version = status.version
    systemInfo.value.startTime = formatTime(status.start_time)

    // Load tasks
    const tasksData = await api.listTasks()
    recentTasks.value = tasksData.tasks ? tasksData.tasks.slice(0, 5) : []
    stats.value.taskCount = tasksData.tasks ? tasksData.tasks.length : 0

    // Load configs
    const configsData = await api.getConfigs()
    stats.value.configCount = configsData.configs ? configsData.configs.length : 0
  } catch (error) {
    console.error('Failed to load data:', error)
    ElMessage.error('加载数据失败')
  }
}

const handleGenerate = async () => {
  generating.value = true
  generateResult.value = null

  try {
    const result = await api.generate({ mode: 'incremental' })
    generateResult.value = {
      type: 'success',
      message: `任务已启动，任务ID: ${result.task_id}`
    }
    setTimeout(() => loadData(), 2000)
  } catch (error) {
    generateResult.value = {
      type: 'error',
      message: '启动任务失败'
    }
  } finally {
    generating.value = false
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

onMounted(() => {
  loadData()
  // Auto refresh every 30s
  setInterval(loadData, 30000)
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

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
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
