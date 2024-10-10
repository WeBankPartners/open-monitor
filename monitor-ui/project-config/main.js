import 'babel-polyfill'
import Vue from 'vue'
import App from '@/App.vue'
import router from './router'
import store from '@/store.js'
import 'bootstrap/dist/js/bootstrap.min.js'
import 'font-awesome/css/font-awesome.css'
import '@/plugins/iview.js'
import httpRequestEntrance from '@/assets/js/httpRequestEntrance.js'
import jquery from 'jquery'
import { tableUtil } from '@/assets/js/tableUtil.js'
import { validate } from '@/assets/js/validate.js'
import VeeValidate from '@/assets/veeValidate/VeeValidate'
import apiCenter from '@/assets/config/api-center.json'
// 引用wecube公共组件和样式
import commonUI from 'wecube-common-ui'
import 'wecube-common-ui/lib/wecube-common-ui.css'
Vue.use(commonUI)

const eventBus = new Vue()
Vue.prototype.$eventBus = eventBus
Vue.prototype.$httpRequestEntrance = httpRequestEntrance
Vue.prototype.JQ = jquery
Vue.prototype.$validate = validate
Vue.prototype.$tableUtil = tableUtil
Vue.prototype.apiCenter = apiCenter

import Title from '@/components/title'
import PageTable from '@/components/table-page/page'
import ModalComponent from '@/components/modal'
import TagShow from '@/components/Tag-show'
Vue.component('Title', Title)
Vue.component('PageTable', PageTable)
Vue.component('ModalComponent', ModalComponent)
Vue.component('TagShow', TagShow)
Vue.use(VeeValidate)

Vue.config.productionTip = false
Vue.config.devtools = true

import DelConfirm from '@/components/del-confirm/index.js'
Vue.use(DelConfirm)

import VueI18n from 'vue-i18n'
import en from 'view-design/dist/locale/en-US'
import zh from 'view-design/dist/locale/zh-CN'

import en_local from '@/assets/locale/lang/en.json'
import zh_local from '@/assets/locale/lang/zh-CN.json'
Vue.use(VueI18n)
Vue.locale = () => {}
const messages = {
  'en-US': Object.assign(en_local, en),
  'zh-CN': Object.assign(zh_local, zh)
}
const i18n = new VueI18n({
  locale:
    localStorage.getItem('lang') ||
    (navigator.language || navigator.userLanguage === 'zh-CN'
      ? 'zh-CN'
      : 'en-US'),
  messages
})

Vue.component('tdSlot', {
  render(createElement) {
    function deepClone(vnodes, createElement) {
      function cloneVNode(vnode) {
        const clonedChildren =
          vnode.children && vnode.children.map(vnode => cloneVNode(vnode))
        const cloned = createElement(vnode.tag, vnode.data, clonedChildren)
        cloned.text = vnode.text
        cloned.isComment = vnode.isComment
        cloned.componentOptions = vnode.componentOptions
        cloned.elm = vnode.elm
        cloned.context = vnode.context
        cloned.ns = vnode.ns
        cloned.isStatic = vnode.isStatic
        cloned.key = vnode.key

        return cloned
      }
      const clonedVNodes = vnodes.map(vnode => cloneVNode(vnode))
      return clonedVNodes
    }
    var slots = this.$parent.$slots.default
    var slot = null
    for (let i = 0; i < slots.length; i++) {
      if (slots[i].data && this.name === slots[i].data.slot) {
        slot = slots[i]
        break
      }
    }
    return createElement(
      'div',
      { class: 'tdslot' },
      deepClone([slot], createElement)
    )
  },
  props: {
    name: {
      type: String,
      default: ''
    }
  }
})

new Vue({
  render: h => h(App),
  store,
  router,
  i18n
}).$mount('#app')
