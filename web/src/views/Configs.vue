<template>
  <div class="configs">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>配置列表</span>
          <div>
            <el-button type="primary" @click="handleGenerateAll" :loading="generating">
              <el-icon><VideoPlay /></el-icon>
              全部生成
            </el-button>
            <el-button @click="loadConfigs">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
        </div>
      </template>

      <el-table
        :data="configs"
        v-loading="loading"
        style="width: 100%"
      >
        <el-table-column prop="name" label="配置名称" width="200" />

        <el-table-column prop="source" label="源路径（Alist）" min-width="250">
          <template #default="scope">
            <el-text type="primary">{{ scope.row.source }}</el-text>
          </template>
        </el-table-column>

        <el-table-column prop="target" label="目标路径（STRM）" min-width="250">
          <template #default="scope">
            <el-text type="success">{{ scope.row.target }}</el-text>
          </template>
        </el-table-column>

        <el-table-column prop="mode" label="更新模式" width="120">
          <template #default="scope">
            <el-tag :type="scope.row.mode === 'full' ? 'warning' : ''" size="small">
              {{ scope.row.mode === 'incremental' ? '增量' : '全量' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="enabled" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.enabled ? 'success' : 'info'" size="small">
              {{ scope.row.enabled ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="200">
          <template #default="scope">
            <el-button
              type="primary"
              size="small"
              :disabled="!scope.row.enabled"
              @click="handleGenerate(scope.row)"
              :loading="generatingMap[scope.row.name]"
            >
              生成
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-alert
        v-if="!configs.length && !loading"
        title="暂无配置"
        type="info"
        description="请在配置文件中添加路径映射配置"
        :closable="false"
        style="margin-top: 20px;"
      />
    </el-card>

    <el-card style="margin-top: 20px;">
      <template #header>
        <span>配置说明</span>
      </template>
      <el-descriptions :column="1" border>
        <el-descriptions-item label="配置文件位置">
          config.yaml
        </el-descriptions-item>
        <el-descriptions-item label="增量模式">
          只处理新增或修改的文件，速度快
        </el-descriptions-item>
        <el-descriptions-item label="全量模式">
          清空目标目录后重新生成所有文件，用于数据修复
        </el-descriptions-item>
        <el-descriptions-item label="配置示例">
          <pre style="margin: 0; padding: 10px; background: #f5f7fa; border-radius: 4px;">
mappings:
  - name: "电影"
    source: "/media/movies"
    target: "/mnt/strm/movies"
    mode: "incremental"
    enabled: true</pre>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../api'

const configs = ref([])
const loading = ref(false)
const generating = ref(false)
const generatingMap = reactive({})

const loadConfigs = async () => {
  loading.value = true
  try {
    const data = await api.getConfigs()
    configs.value = data.configs || []
  } catch (error) {
    console.error('Failed to load configs:', error)
    ElMessage.error('加载配置失败')
  } finally {
    loading.value = false
  }
}

const handleGenerate = async (config) => {
  generatingMap[config.name] = true

  try {
    const result = await api.generate({
      path: config.name,
      mode: config.mode
    })
    ElMessage.success(`任务已启动：${result.task_id}`)
  } catch (error) {
    ElMessage.error('启动任务失败')
  } finally {
    generatingMap[config.name] = false
  }
}

const handleGenerateAll = async () => {
  generating.value = true

  try {
    const result = await api.generate({
      mode: 'incremental'
    })
    ElMessage.success(`任务已启动：${result.task_id}`)
  } catch (error) {
    ElMessage.error('启动任务失败')
  } finally {
    generating.value = false
  }
}

onMounted(() => {
  loadConfigs()
})
</script>

<style scoped>
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

pre {
  font-family: 'Courier New', Courier, monospace;
  font-size: 12px;
  color: #606266;
}
</style>
