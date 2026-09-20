import { createApp } from 'vue'
import '@fontsource/roboto/400.css'
import '@fontsource/roboto/500.css'
import '@fontsource/roboto/700.css'
import './theme'
import App from './App.vue'
import router from './router'
import './style.css'

createApp(App).use(router).mount('#app')
