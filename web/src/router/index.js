import { createRouter, createWebHistory } from 'vue-router'
import Dashboard from '../views/Dashboard.vue'
import Tasks from '../views/Tasks.vue'
import Configs from '../views/Configs.vue'

const routes = [
  {
    path: '/',
    name: 'Dashboard',
    component: Dashboard
  },
  {
    path: '/tasks',
    name: 'Tasks',
    component: Tasks
  },
  {
    path: '/configs',
    name: 'Configs',
    component: Configs
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
