<template>
  <div class="configs">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>配置列表</span>
          <div>
            <el-button type="success" @click="showAddDialog">
              <el-icon><Plus /></el-icon>
              新增配置
            </el-button>
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

        <el-table-column label="操作" width="250">
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
            <el-button
              type="warning"
              size="small"
              @click="showEditDialog(scope.row)"
            >
              编辑
            </el-button>
            <el-button
              type="danger"
              size="small"
              @click="handleDelete(scope.row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-alert
        v-if="!configs.length && !loading"
        title="暂无配置"
        type="info"
        description="请点击「新增配置」按钮添加路径映射配置"
        :closable="false"
        style="margin-top: 20px;"
      />
    </el-card>

    <!-- 新增/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'add' ? '新增配置' : '编辑配置'"
      width="600px"
    >
      <el-form :model="formData" :rules="formRules" ref="formRef" label-width="120px">
        <el-form-item label="配置名称" prop="name">
          <el-input v-model="formData.name" placeholder="例如: 电影" />
        </el-form-item>

        <el-form-item label="源路径" prop="source">
          <el-input v-model="formData.source" placeholder="例如: /media/movies">
            <template #prepend>Alist</template>
          </el-input>
        </el-form-item>

        <el-form-item label="目标路径" prop="target">
          <el-input v-model="formData.target" placeholder="例如: /mnt/strm/movies">
            <template #prepend>STRM</template>
          </el-input>
        </el-form-item>

        <el-form-item label="更新模式" prop="mode">
          <el-radio-group v-model="formData.mode">
            <el-radio value="incremental">增量模式（只处理新增文件）</el-radio>
            <el-radio value="full">全量模式（清空后重新生成）</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="启用状态" prop="enabled">
          <el-switch v-model="formData.enabled" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          {{ dialogMode === 'add' ? '创建' : '保存' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, VideoPlay, Refresh } from '@element-plus/icons-vue'
import api from '../api'

const configs = ref([])
const loading = ref(false)
const generating = ref(false)
const generatingMap = reactive({})
const dialogVisible = ref(false)
const dialogMode = ref('add') // 'add' or 'edit'
const submitting = ref(false)
const formRef = ref(null)

const formData = reactive({
  id: null,
  name: '',
  source: '',
  target: '',
  mode: 'incremental',
  enabled: true
})

const formRules = {
  name: [
    { required: true, message: '请输入配置名称', trigger: 'blur' }
  ],
  source: [
    { required: true, message: '请输入源路径', trigger: 'blur' }
  ],
  target: [
    { required: true, message: '请输入目标路径', trigger: 'blur' }
  ],
  mode: [
    { required: true, message: '请选择更新模式', trigger: 'change' }
  ]
}

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

const showAddDialog = () => {
  dialogMode.value = 'add'
  resetForm()
  dialogVisible.value = true
}

const showEditDialog = (config) => {
  dialogMode.value = 'edit'
  formData.id = config.id
  formData.name = config.name
  formData.source = config.source
  formData.target = config.target
  formData.mode = config.mode
  formData.enabled = config.enabled
  dialogVisible.value = true
}

const resetForm = () => {
  formData.id = null
  formData.name = ''
  formData.source = ''
  formData.target = ''
  formData.mode = 'incremental'
  formData.enabled = true
  if (formRef.value) {
    formRef.value.clearValidate()
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
  } catch {
    return
  }

  submitting.value = true

  try {
    const data = {
      name: formData.name,
      source: formData.source,
      target: formData.target,
      mode: formData.mode,
      enabled: formData.enabled
    }

    if (dialogMode.value === 'add') {
      await api.createConfig(data)
      ElMessage.success('配置创建成功')
    } else {
      await api.updateConfig(formData.id, data)
      ElMessage.success('配置更新成功')
    }

    dialogVisible.value = false
    loadConfigs()
  } catch (error) {
    console.error('Failed to save config:', error)
    ElMessage.error(dialogMode.value === 'add' ? '创建配置失败' : '更新配置失败')
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (config) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除配置 "${config.name}" 吗？此操作不可恢复。`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await api.deleteConfig(config.id)
    ElMessage.success('配置删除成功')
    loadConfigs()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('Failed to delete config:', error)
      ElMessage.error('删除配置失败')
    }
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
</style>
