import Vue from 'vue'
import Index from './index.vue'

let messageInstance = null
const MessageConstructor = Vue.extend(Index)

const init = () => {
  messageInstance = new MessageConstructor()
  messageInstance.$mount()
  document.body.appendChild(messageInstance.$el)
}

const caller = options => {
  if (!messageInstance) {
    init(options)
  }
  messageInstance.add(options)
}

export default {
  // 返回 install 函数 用于 Vue.use 注册
  install(vue) {
    vue.prototype.$delConfirm = caller
  }
}
