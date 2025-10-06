<template>
  <div class="tasks">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>任务列表</span>
          <el-button type="primary" @click="loadTasks">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
        </div>
      </template>

      <el-table
        :data="tasks"
        v-loading="loading"
        style="width: 100%"
      >
        <el-table-column prop="task_id" label="任务ID" width="280">
          <template #default="scope">
            <el-text type="info" size="small">{{ scope.row.task_id }}</el-text>
          </template>
        </el-table-column>

        <el-table-column prop="config_name" label="配置名称" width="150" />

        <el-table-column prop="mode" label="模式" width="100">
          <template #default="scope">
            <el-tag size="small" :type="scope.row.mode === 'full' ? 'warning' : ''">
              {{ getModeText(scope.row.mode) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="status" label="状态" width="120">
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

        <el-table-column label="操作" width="150">
          <template #default="scope">
            <el-button
              type="text"
              size="small"
              @click="showTaskDetail(scope.row)"
            >
              查看详情
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-container">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <!-- Task Detail Dialog -->
    <el-dialog
      v-model="detailVisible"
      title="任务详情"
      width="600px"
    >
      <el-descriptions :column="1" border v-if="currentTask">
        <el-descriptions-item label="任务ID">{{ currentTask.task_id }}</el-descriptions-item>
        <el-descriptions-item label="配置名称">{{ currentTask.config_name }}</el-descriptions-item>
        <el-descriptions-item label="模式">
          <el-tag :type="currentTask.mode === 'full' ? 'warning' : ''">
            {{ getModeText(currentTask.mode) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="getStatusType(currentTask.status)">
            {{ getStatusText(currentTask.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="创建文件">{{ currentTask.files_created }}</el-descriptions-item>
        <el-descriptions-item label="删除文件">{{ currentTask.files_deleted }}</el-descriptions-item>
        <el-descriptions-item label="跳过文件">{{ currentTask.files_skipped }}</el-descriptions-item>
        <el-descriptions-item label="开始时间">{{ formatTime(currentTask.started_at) }}</el-descriptions-item>
        <el-descriptions-item label="完成时间">{{ formatTime(currentTask.completed_at) }}</el-descriptions-item>
        <el-descriptions-item label="错误信息" v-if="currentTask.errors">
          <el-text type="danger">{{ currentTask.errors }}</el-text>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../api'

const tasks = ref([])
const loading = ref(false)
const detailVisible = ref(false)
const currentTask = ref(null)
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)

const loadTasks = async () => {
  loading.value = true
  try {
    const data = await api.listTasks(currentPage.value, pageSize.value)
    tasks.value = data.tasks || []
    total.value = data.total || 0
  } catch (error) {
    console.error('Failed to load tasks:', error)
    ElMessage.error('加载任务列表失败')
  } finally {
    loading.value = false
  }
}

const handlePageChange = (page) => {
  currentPage.value = page
  loadTasks()
}

const handleSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1
  loadTasks()
}

const showTaskDetail = (task) => {
  currentTask.value = task
  detailVisible.value = true
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

const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
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
  loadTasks()
})
</script>

<style scoped>
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

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style>
