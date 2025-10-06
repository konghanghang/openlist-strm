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

        <el-table-column prop="extensions" label="扩展名" width="150">
          <template #default="scope">
            <el-text size="small">{{ scope.row.extensions?.join(', ') }}</el-text>
          </template>
        </el-table-column>

        <el-table-column prop="mode" label="更新模式" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.mode === 'full' ? 'warning' : ''" size="small">
              {{ scope.row.mode === 'incremental' ? '增量' : '全量' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="strm_mode" label="STRM模式" width="120">
          <template #default="scope">
            <el-tag :type="scope.row.strm_mode === 'http_url' ? 'danger' : 'primary'" size="small">
              {{ scope.row.strm_mode === 'alist_path' ? '路径' : '直链' }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column prop="cron_expr" label="定时任务" width="150">
          <template #default="scope">
            <el-text v-if="scope.row.cron_expr" type="success" size="small">
              {{ scope.row.cron_expr }}
            </el-text>
            <el-text v-else type="info" size="small">未设置</el-text>
          </template>
        </el-table-column>

        <el-table-column prop="enabled" label="状态" width="80">
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

        <el-form-item label="视频扩展名" prop="extensions">
          <el-select
            v-model="formData.extensions"
            multiple
            placeholder="选择视频文件扩展名"
            style="width: 100%"
          >
            <el-option label="mp4" value="mp4" />
            <el-option label="mkv" value="mkv" />
            <el-option label="avi" value="avi" />
            <el-option label="mov" value="mov" />
            <el-option label="flv" value="flv" />
            <el-option label="wmv" value="wmv" />
            <el-option label="ts" value="ts" />
            <el-option label="m4v" value="m4v" />
          </el-select>
        </el-form-item>

        <el-form-item label="并发数" prop="concurrent">
          <el-input-number v-model="formData.concurrent" :min="1" :max="20" style="width: 100%" />
          <div style="color: #909399; font-size: 12px; margin-top: 5px;">
            同时处理的文件数量，建议 1-5，过大可能触发网盘风控
          </div>
        </el-form-item>

        <el-form-item label="更新模式" prop="mode">
          <el-radio-group v-model="formData.mode">
            <el-radio value="incremental">增量模式（只处理新增文件）</el-radio>
            <el-radio value="full">全量模式（清空后重新生成）</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="STRM内容" prop="strm_mode">
          <el-radio-group v-model="formData.strm_mode">
            <el-radio value="alist_path">Alist路径（配合MediaWarp使用）</el-radio>
            <el-radio value="http_url">直链URL（直接播放）</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item label="定时任务" prop="cron_expr">
          <div style="width: 100%;">
            <!-- 配置控件 -->
            <div style="display: flex; gap: 10px; align-items: center; flex-wrap: wrap;">
              <el-select
                v-model="cronMode"
                placeholder="选择频率"
                style="width: 140px"
                @change="handleCronModeChange"
              >
                <el-option label="不启用" value="disabled" />
                <el-option label="每隔N分钟" value="interval_minutes" />
                <el-option label="每小时" value="hourly" />
                <el-option label="每天" value="daily" />
                <el-option label="每周" value="weekly" />
                <el-option label="每月" value="monthly" />
                <el-option label="自定义表达式" value="custom" />
              </el-select>

              <!-- 每隔N分钟 -->
              <template v-if="cronMode === 'interval_minutes'">
                <span>每</span>
                <el-select v-model="cronIntervalMinutes" style="width: 100px" @change="updateCronExpr">
                  <el-option label="5分钟" :value="5" />
                  <el-option label="10分钟" :value="10" />
                  <el-option label="15分钟" :value="15" />
                  <el-option label="20分钟" :value="20" />
                  <el-option label="30分钟" :value="30" />
                </el-select>
                <span>执行一次</span>
              </template>

              <!-- 每小时 -->
              <template v-if="cronMode === 'hourly'">
                <span>每小时的第</span>
                <el-input-number v-model="cronMinute" :min="0" :max="59" style="width: 100px" @change="updateCronExpr" />
                <span>分钟</span>
              </template>

              <!-- 每天 -->
              <template v-if="cronMode === 'daily'">
                <span>每天</span>
                <el-time-select
                  v-model="cronDailyTime"
                  start="00:00"
                  step="00:30"
                  end="23:30"
                  placeholder="选择时间"
                  style="width: 120px"
                  @change="updateCronExpr"
                />
              </template>

              <!-- 每周 -->
              <template v-if="cronMode === 'weekly'">
                <span>每周</span>
                <el-select v-model="cronWeekday" style="width: 100px" @change="updateCronExpr">
                  <el-option label="周一" :value="1" />
                  <el-option label="周二" :value="2" />
                  <el-option label="周三" :value="3" />
                  <el-option label="周四" :value="4" />
                  <el-option label="周五" :value="5" />
                  <el-option label="周六" :value="6" />
                  <el-option label="周日" :value="0" />
                </el-select>
                <el-time-select
                  v-model="cronWeeklyTime"
                  start="00:00"
                  step="00:30"
                  end="23:30"
                  placeholder="选择时间"
                  style="width: 120px"
                  @change="updateCronExpr"
                />
              </template>

              <!-- 每月 -->
              <template v-if="cronMode === 'monthly'">
                <span>每月</span>
                <el-input-number v-model="cronMonthDay" :min="1" :max="31" style="width: 100px" @change="updateCronExpr" />
                <span>号</span>
                <el-time-select
                  v-model="cronMonthlyTime"
                  start="00:00"
                  step="00:30"
                  end="23:30"
                  placeholder="选择时间"
                  style="width: 120px"
                  @change="updateCronExpr"
                />
              </template>

              <!-- 自定义 -->
              <el-input
                v-if="cronMode === 'custom'"
                v-model="formData.cron_expr"
                placeholder="如: 0 2 * * *"
                style="max-width: 300px"
              />
            </div>

            <!-- 表达式和执行时间预览 -->
            <div style="margin-top: 10px; color: #909399; font-size: 12px;">
              <div v-if="formData.cron_expr">
                <div style="margin-bottom: 5px;">
                  Cron 表达式：<el-text type="success">{{ formData.cron_expr }}</el-text>
                </div>
                <div v-if="nextRunTimes.length > 0">
                  <div style="margin-bottom: 3px;">最近三次执行时间：</div>
                  <div v-for="(time, index) in nextRunTimes" :key="index" style="margin-left: 10px; line-height: 1.8;">
                    <el-text type="warning">{{ index + 1 }}. {{ time }}</el-text>
                  </div>
                </div>
              </div>
              <div v-else style="color: #909399;">
                留空表示不启用定时任务
              </div>
            </div>
          </div>
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
  extensions: ['mp4', 'mkv', 'avi', 'mov'],
  concurrent: 3,
  mode: 'incremental',
  strm_mode: 'alist_path',
  cron_expr: '',
  enabled: true
})

// Cron 配置相关
const cronMode = ref('disabled')
const cronIntervalMinutes = ref(30)
const cronMinute = ref(0)
const cronDailyTime = ref('02:00')
const cronWeekday = ref(0)
const cronWeeklyTime = ref('02:00')
const cronMonthDay = ref(1)
const cronMonthlyTime = ref('02:00')
const nextRunTimes = ref([])

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
  extensions: [
    { required: true, type: 'array', min: 1, message: '请至少选择一个扩展名', trigger: 'change' }
  ],
  concurrent: [
    { required: true, type: 'number', min: 1, max: 20, message: '并发数必须在 1-20 之间', trigger: 'change' }
  ],
  mode: [
    { required: true, message: '请选择更新模式', trigger: 'change' }
  ],
  strm_mode: [
    { required: true, message: '请选择STRM内容模式', trigger: 'change' }
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
  formData.extensions = config.extensions || ['mp4', 'mkv', 'avi', 'mov']
  formData.concurrent = config.concurrent || 10
  formData.mode = config.mode
  formData.strm_mode = config.strm_mode || 'alist_path'
  formData.cron_expr = config.cron_expr || ''
  formData.enabled = config.enabled
  parseCronExpr(config.cron_expr || '')
  dialogVisible.value = true
}

// 解析 Cron 表达式到 UI 控件
const parseCronExpr = (expr) => {
  if (!expr) {
    cronMode.value = 'disabled'
    return
  }

  const parts = expr.split(' ')
  // 支持 5 字段（分钟级）和 6 字段（秒级）格式
  if (parts.length !== 5 && parts.length !== 6) {
    cronMode.value = 'custom'
    return
  }

  // 6 字段格式: 秒 分 时 日 月 周
  // 5 字段格式: 分 时 日 月 周（兼容旧数据）
  let second, minute, hour, day, month, weekday
  if (parts.length === 6) {
    [second, minute, hour, day, month, weekday] = parts
  } else {
    [minute, hour, day, month, weekday] = parts
    second = '0'
  }

  // 每隔N分钟: 0 */5 * * * *
  if (second === '0' && minute.startsWith('*/') && hour === '*' && day === '*' && month === '*' && weekday === '*') {
    cronMode.value = 'interval_minutes'
    cronIntervalMinutes.value = parseInt(minute.substring(2))
    return
  }

  // 每小时: 0 0 * * * *
  if (second === '0' && hour === '*' && day === '*' && month === '*' && weekday === '*') {
    cronMode.value = 'hourly'
    cronMinute.value = parseInt(minute)
    return
  }

  // 每天: 0 0 2 * * *
  if (second === '0' && day === '*' && month === '*' && weekday === '*') {
    cronMode.value = 'daily'
    cronDailyTime.value = `${hour.padStart(2, '0')}:${minute.padStart(2, '0')}`
    return
  }

  // 每周: 0 0 2 * * 0
  if (second === '0' && day === '*' && month === '*' && weekday !== '*') {
    cronMode.value = 'weekly'
    cronWeekday.value = parseInt(weekday)
    cronWeeklyTime.value = `${hour.padStart(2, '0')}:${minute.padStart(2, '0')}`
    return
  }

  // 每月: 0 0 2 1 * *
  if (second === '0' && month === '*' && weekday === '*' && day !== '*') {
    cronMode.value = 'monthly'
    cronMonthDay.value = parseInt(day)
    cronMonthlyTime.value = `${hour.padStart(2, '0')}:${minute.padStart(2, '0')}`
    return
  }

  // 其他情况当作自定义
  cronMode.value = 'custom'
}

// 模式切换时的处理
const handleCronModeChange = () => {
  if (cronMode.value === 'disabled') {
    formData.cron_expr = ''
    nextRunTimes.value = []
  } else {
    updateCronExpr()
  }
}

// 更新 Cron 表达式（6 字段格式：秒 分 时 日 月 周）
const updateCronExpr = () => {
  switch (cronMode.value) {
    case 'disabled':
      formData.cron_expr = ''
      break
    case 'interval_minutes':
      formData.cron_expr = `0 */${cronIntervalMinutes.value} * * * *`
      break
    case 'hourly':
      formData.cron_expr = `0 ${cronMinute.value} * * * *`
      break
    case 'daily': {
      const [h, m] = cronDailyTime.value.split(':')
      formData.cron_expr = `0 ${parseInt(m)} ${parseInt(h)} * * *`
      break
    }
    case 'weekly': {
      const [h, m] = cronWeeklyTime.value.split(':')
      formData.cron_expr = `0 ${parseInt(m)} ${parseInt(h)} * * ${cronWeekday.value}`
      break
    }
    case 'monthly': {
      const [h, m] = cronMonthlyTime.value.split(':')
      formData.cron_expr = `0 ${parseInt(m)} ${parseInt(h)} ${cronMonthDay.value} * *`
      break
    }
  }
  calculateNextRunTime()
}

// 计算最近三次执行时间
const calculateNextRunTime = () => {
  if (!formData.cron_expr) {
    nextRunTimes.value = []
    return
  }

  const now = new Date()
  const times = []

  try {
    // 计算三次执行时间
    for (let i = 0; i < 3; i++) {
      let next = new Date(i === 0 ? now : times[i - 1].date)

      switch (cronMode.value) {
        case 'interval_minutes':
          if (i === 0) {
            next.setMinutes(now.getMinutes() + cronIntervalMinutes.value)
          } else {
            next.setMinutes(next.getMinutes() + cronIntervalMinutes.value)
          }
          break
        case 'hourly':
          if (i === 0) {
            next.setHours(now.getHours() + 1)
            next.setMinutes(cronMinute.value)
            next.setSeconds(0)
            if (next <= now) next.setHours(next.getHours() + 1)
          } else {
            next.setHours(next.getHours() + 1)
          }
          break
        case 'daily': {
          const [h, m] = cronDailyTime.value.split(':')
          if (i === 0) {
            next.setHours(parseInt(h), parseInt(m), 0)
            if (next <= now) next.setDate(next.getDate() + 1)
          } else {
            next.setDate(next.getDate() + 1)
          }
          break
        }
        case 'weekly': {
          const [h, m] = cronWeeklyTime.value.split(':')
          const targetDay = cronWeekday.value
          if (i === 0) {
            const currentDay = now.getDay()
            let daysUntil = targetDay - currentDay
            if (daysUntil <= 0) daysUntil += 7
            next.setDate(now.getDate() + daysUntil)
            next.setHours(parseInt(h), parseInt(m), 0)
          } else {
            next.setDate(next.getDate() + 7)
          }
          break
        }
        case 'monthly': {
          const [h, m] = cronMonthlyTime.value.split(':')
          if (i === 0) {
            next.setDate(cronMonthDay.value)
            next.setHours(parseInt(h), parseInt(m), 0)
            if (next <= now) next.setMonth(next.getMonth() + 1)
          } else {
            next.setMonth(next.getMonth() + 1)
          }
          break
        }
        default:
          nextRunTimes.value = []
          return
      }

      times.push({
        date: new Date(next),
        formatted: next.toLocaleString('zh-CN', {
          year: 'numeric',
          month: '2-digit',
          day: '2-digit',
          hour: '2-digit',
          minute: '2-digit'
        })
      })
    }

    nextRunTimes.value = times.map(t => t.formatted)
  } catch (e) {
    nextRunTimes.value = []
  }
}

const resetForm = () => {
  formData.id = null
  formData.name = ''
  formData.source = ''
  formData.target = ''
  formData.extensions = ['mp4', 'mkv', 'avi', 'mov']
  formData.concurrent = 3
  formData.mode = 'incremental'
  formData.strm_mode = 'alist_path'
  formData.cron_expr = ''
  formData.enabled = true

  // Reset cron fields
  cronMode.value = 'disabled'
  cronIntervalMinutes.value = 30
  cronMinute.value = 0
  cronDailyTime.value = '02:00'
  cronWeekday.value = 0
  cronWeeklyTime.value = '02:00'
  cronMonthDay.value = 1
  cronMonthlyTime.value = '02:00'
  nextRunTimes.value = []

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
      extensions: formData.extensions,
      concurrent: formData.concurrent,
      mode: formData.mode,
      strm_mode: formData.strm_mode,
      cron_expr: formData.cron_expr,
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
