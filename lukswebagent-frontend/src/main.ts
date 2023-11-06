import './assets/background.css'
import 'mdui/dist/css/mdui.css'
import { createApp } from 'vue'
import { createI18n } from 'vue-i18n'
import App from './App.vue'
const messages = {
  en: {
    title: "The server is under maintenance",
    tips: "Please wait for administrators to unlock the server",
    password: "Password",
    unlock: "Unlock Server",
    unlocking: "Unlocking",
    unlocked: "Unlocked",
    waitserver:"Waiting for the server to start",
    redirecting:"Redirecting...",
    errnetwork:"Network Error",
    code: {
      400: "Bad request",
      401: "Authenticate failed"
    }
  },
  cn: {
    title: "该服务器正在维护",
    tips: "请等待管理人员解锁服务器",
    password: "解锁密码",
    unlock: "解锁服务器",
    unlocking: "正在解锁",
    unlocked: "解锁成功",
    waitserver:"等待服务器启动....",
    redirecting:"正在进入服务器",
    errnetwork:"网络错误",
    code: {
      400: "错误的请求",
      401: "认证失败"
    }
  }
}
const i18n = createI18n({
  legacy: false,
  locale: 'en',
  fallbackLocale: 'en',
  messages
})
const app = createApp(App)
app.use(i18n)
app.mount('#app')
