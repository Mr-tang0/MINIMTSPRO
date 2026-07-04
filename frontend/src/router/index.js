import { createRouter, createWebHistory } from 'vue-router'
import Login from '../components/Login.vue'
import Update from '../components/Update.vue'
import MINIMTS from '../components/MINIMTS.vue'
import Camera from '../components/Camera.vue'
import Project from '../components/Project.vue'
import ROISelector from '../components/ROISelector.vue'
import System from '../components/System.vue'

const routes = [
  {
    path: '/',
    name: 'Login',
    component: Login
  },
  {
    path: '/update',
    name: 'Update',
    component: Update
  },
  {
    path: '/mts',
    name: 'MTS',
    component: MINIMTS
  },
  {
    path: '/camera',
    name: 'Camera',
    component: Camera
  },
  {
    path: '/system',
    name: 'System',
    component: System
  },
  {
    path: '/project',
    name: 'Project',
    component: Project
  },
  {
    path: '/roi-selector',
    name: 'ROISelector',
    component: ROISelector
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router