import { createApp } from 'vue'
import '@tabler/core/dist/css/tabler.min.css'
import './theme'
import App from './App.vue'
import router from './router'
import './style.css'

createApp(App).use(router).mount('#app')
