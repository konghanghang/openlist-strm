import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 30000
})

// Request interceptor
api.interceptors.request.use(
  config => {
    // Add API token if configured
    const token = localStorage.getItem('api_token')
    if (token) {
      config.headers['X-API-Token'] = token
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// Response interceptor
api.interceptors.response.use(
  response => response.data,
  error => {
    console.error('API Error:', error)
    return Promise.reject(error)
  }
)

export default {
  // Health check
  health() {
    return axios.get('/health')
  },

  // Generate STRM files
  generate(data) {
    return api.post('/generate', data)
  },

  // Get task by ID
  getTask(taskId) {
    return api.get(`/tasks/${taskId}`)
  },

  // List tasks with pagination
  listTasks(page = 1, pageSize = 20) {
    return api.get('/tasks', {
      params: { page, page_size: pageSize }
    })
  },

  // Get configs
  getConfigs() {
    return api.get('/configs')
  },

  // Create config/mapping
  createConfig(data) {
    return api.post('/configs', data)
  },

  // Update config/mapping
  updateConfig(id, data) {
    return api.put(`/configs/${id}`, data)
  },

  // Delete config/mapping
  deleteConfig(id) {
    return api.delete(`/configs/${id}`)
  },

  // Get system status
  getStatus() {
    return api.get('/status')
  }
}
