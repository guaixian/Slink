import { createRouter, createWebHistory } from 'vue-router'
import Slink from '../components/Slink.vue'
import UserHome from '../components/UserHome.vue'
import InitSetup from '../components/InitSetup.vue'

import UploadPage from '../components/user/UploadPage.vue'
import WatermarkPage from '../components/user/WatermarkPage.vue'
import ImageToolsPage from '../components/user/ImageToolsPage.vue'
import AIToolsPage from '../components/user/AIToolsPage.vue'
import Dashboard from '../components/user/Dashboard.vue'
import MyImages from '../components/user/MyImages.vue'
import BasicSettings from '../components/user/BasicSettings.vue'
import ApiReference from '../components/user/ApiReference.vue'

import ImageManagement from '../components/admin/ImageManagement.vue'
import SystemConsole from '../components/admin/SystemConsole.vue'
import SystemSettings from '../components/admin/SystemSettings.vue'
import UploadPolicySettings from '../components/admin/UploadPolicySettings.vue'
import StorageStrategy from '../components/admin/StorageStrategy.vue'
import StorageStrategyEdit from '../components/admin/StorageStrategyEdit.vue'

import { setupRouterGuards } from './guards'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'index',
      component: Slink,
    },
    {
      path: '/init',
      name: 'init-setup',
      component: InitSetup,
    },
    {
      path: '/admin',
      name: 'admin',
      component: UserHome,
      children: [
        {
          path: '',
          name: 'admin-upload',
          component: UploadPage,
        },
        {
          path: 'watermark',
          name: 'admin-watermark',
          component: WatermarkPage,
        },
        {
          path: 'image-tools',
          name: 'admin-image-tools',
          component: ImageToolsPage,
        },
        {
          path: 'ai-tools',
          name: 'admin-ai-tools',
          component: AIToolsPage,
        },
        {
          path: 'dashboard',
          name: 'admin-dashboard',
          component: Dashboard,
        },
        {
          path: 'images',
          name: 'admin-images',
          component: MyImages,
        },
        {
          path: 'basic-settings',
          name: 'admin-basic-settings',
          component: BasicSettings,
        },
        {
          path: 'api-reference',
          name: 'admin-api-reference',
          component: ApiReference,
        },
        {
          path: 'image-management',
          name: 'admin-image-management',
          component: ImageManagement,
        },
        {
          path: 'console',
          name: 'admin-console',
          component: SystemConsole,
        },
        {
          path: 'storage',
          name: 'admin-storage',
          component: StorageStrategy,
        },
        {
          path: 'storage/create',
          name: 'admin-storage-create',
          component: StorageStrategyEdit,
        },
        {
          path: 'storage/edit/:id',
          name: 'admin-storage-edit',
          component: StorageStrategyEdit,
        },
        {
          path: 'upload-policy',
          name: 'admin-upload-policy',
          component: UploadPolicySettings,
        },
        {
          path: 'settings',
          name: 'admin-settings',
          component: SystemSettings,
        },
      ],
    },
  ],
})

setupRouterGuards(router)

export default router
