import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { useUserStore } from './stores/user'
import messagePlugin from './plugins/message'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.use(messagePlugin)

// 在应用挂载前初始化token
const userStore = useUserStore()
userStore.initUserFromStorage()

app.mount('#app')
