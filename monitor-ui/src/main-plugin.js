import store from './store'
import router from './router-plugin'
import '@/assets/css/local.bootstrap.css'
import 'bootstrap/dist/js/bootstrap.min.js'
import 'font-awesome/css/font-awesome.css'
import './plugins/iview.js'
// import '@/assets/locale/language'
import httpRequestEntrance from '@/assets/js/httpRequestEntrance.js'
import jquery from 'jquery'
import {tableUtil} from '@/assets/js/tableUtil.js'
import {validate} from '@/assets/js/validate.js'
import VeeValidate from '@/assets/veeValidate/VeeValidate'
import apiCenter from '@/assets/config/api-center.json'
import vuex from 'vuex'
window.use(vuex)
window.addOptions({
  $httpRequestEntrance: httpRequestEntrance,
  JQ: jquery,
  $store: store,
  $validate: validate,
  $tableUtil: tableUtil,
  apiCenter: apiCenter
})

window.addRoutes(router, 'open-monitor')

import Title from '@/components/title'
import PageTable from '@/components/table-page/page'
import ModalComponent from '@/components/modal'
window.component('Title', Title)
window.component('PageTable', PageTable)
window.component('ModalComponent', ModalComponent)
window.use(VeeValidate)

window.component('tdSlot', {
  render(createElement) {
    function deepClone(vnodes, createElement) {

      function cloneVNode (vnode) {
        const clonedChildren = vnode.children && vnode.children.map(vnode => cloneVNode(vnode))
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
    for(let i=0;i<slots.length;i++){
      if(slots[i].data && this.name === slots[i].data.slot){
        slot = slots[i]
        break
      }
    }
    return createElement('div',{class:'tdslot'},deepClone([slot], createElement))
  },
  props:{
    name:{
      type:String,
      default:''
    }
  }
})


import en_local from '@/assets/locale/lang/en.json';
import zh_local from '@/assets/locale/lang/zh-CN.json';
window.locale('en-US',en_local)
window.locale('zh-CN',zh_local)
