import Vue from 'vue'
import store from '@/store.js'
import router from './router-plugin'
import routerP from './router-plugin-p'
import '@/assets/css/local.bootstrap.css'
import 'bootstrap/dist/js/bootstrap.min.js'
import 'font-awesome/css/font-awesome.css'
import '@/plugins/iview.js'
import httpRequestEntrance from '@/assets/js/httpRequestEntrance.js'
import jquery from 'jquery'
import {tableUtil} from '@/assets/js/tableUtil.js'
import {validate} from '@/assets/js/validate.js'
import VeeValidate from '@/assets/veeValidate/VeeValidate'
import apiCenter from '@/assets/config/api-center.json'
import vuex from 'vuex'
const eventBus = new Vue()

window.use(vuex)
window.addOptions({
  $httpRequestEntrance: httpRequestEntrance,
  JQ: jquery,
  $store: store,
  $validate: validate,
  $tableUtil: tableUtil,
  apiCenter: apiCenter,
  $eventBus: eventBus
})

const implicitRoute = {
  'monitorConfigIndex/endpointManagement': {
    parentBreadcrumb: {'zh-CN': '对象', 'en-US': 'Object'},
    childBreadcrumb: { 'zh-CN': '对象', 'en-US': 'Object' }
  },
  'monitorConfigIndex/groupManagement': {
    parentBreadcrumb: {'zh-CN': '对象', 'en-US': 'Object'},
    childBreadcrumb: { 'zh-CN': '对象组', 'en-US': 'Object Group' }
  },
  'monitorConfigIndex/thresholdManagement': {
    parentBreadcrumb: {'zh-CN': '告警', 'en-US': 'Alarm'},
    childBreadcrumb: { 'zh-CN': '指标阈值', 'en-US': 'Metric Threshold' }
  },
  'monitorConfigIndex/logManagement': {
    parentBreadcrumb: {'zh-CN': '告警', 'en-US': 'Alarm'},
    childBreadcrumb: { 'zh-CN': '关键字', 'en-US': 'Keywords' }
  },
  'monitorConfigIndex/resourceLevel': {
    parentBreadcrumb: {'zh-CN': '对象', 'en-US': 'Object'},
    childBreadcrumb: { 'zh-CN': '层级对象', 'en-US': 'Hierarchy Object' }
  },
  'monitorConfigIndex/exporter': {
    parentBreadcrumb: {'zh-CN': '其他', 'en-US': 'Other'},
    childBreadcrumb: { 'zh-CN': '采集器', 'en-US': 'Collector'}
  },
  'monitorConfigIndex/businessMonitor': {
    parentBreadcrumb: {'zh-CN': '指标123', 'en-US': 'quota'},
    childBreadcrumb: { 'zh-CN': '业务配置', 'en-US': 'Business Configuration' }
  },
  'monitorConfigIndex/logTemplate': {
    parentBreadcrumb: {'zh-CN': '指标', 'en-US': 'quota'},
    childBreadcrumb: { 'zh-CN': '业务日志模板', 'en-US': 'Business Log Template' }
  },
  'monitorConfigIndex/metricConfig': {
    parentBreadcrumb: {'zh-CN': '指标', 'en-US': 'quota'},
    childBreadcrumb: { 'zh-CN': '自定义', 'en-US': 'customize' }
  },
  'monitorConfigIndex/groupBoard': {
    parentBreadcrumb: {'zh-CN': '对象', 'en-US': 'Object'},
    childBreadcrumb: { 'zh-CN': '对象看板', 'en-US': 'Object Board' }
  },
  'viewConfig': {
    parentBreadcrumb: {'zh-CN': '监控', 'en-US': 'Monitor'},
    childBreadcrumb: { 'zh-CN': '查看视图', 'en-US': 'View' }
  },
  'editLineView': {
    parentBreadcrumb: {'zh-CN': '监控', 'en-US': 'Monitor'},
    childBreadcrumb: { 'zh-CN': '编辑视图', 'en-US': 'View' }
  },
  'editPieView': {
    parentBreadcrumb: {'zh-CN': '监控', 'en-US': 'Monitor'},
    childBreadcrumb: { 'zh-CN': '编辑视图', 'en-US': 'View' }
  },
  'viewChart': {
    parentBreadcrumb: {'zh-CN': '监控', 'en-US': 'Monitor'},
    childBreadcrumb: { 'zh-CN': '视图', 'en-US': 'View' }
  },
  'alarmHistory': {
    parentBreadcrumb: {'zh-CN': '监控', 'en-US': 'Monitor'},
    childBreadcrumb: { 'zh-CN': '历史告警', 'en-US': 'History Alarm' }
  }
}
window.addImplicitRoute(implicitRoute)
window.addRoutersWithoutPermission(routerP, 'open-monitor')
window.addRoutes(router, 'open-monitor')

import Title from '@/components/title'
import PageTable from '@/components/table-page/page'
import ModalComponent from '@/components/modal'
window.component('Title', Title)
window.component('PageTable', PageTable)
window.component('ModalComponent', ModalComponent)
window.use(VeeValidate)
import DelConfirm from '@/components/del-confirm/index.js'
window.use(DelConfirm)

import Dashboard from '@/views/dashboard'
window.addHomepageComponent && window.addHomepageComponent({
  code: 'MONITORING',
  name: () => {
    return window.vm.$t('menu.homepageName')
  },
  component: Dashboard
})

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
    for(let i=0; i<slots.length; i++){
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


import en_local from '@/assets/locale/lang/en.json'
import zh_local from '@/assets/locale/lang/zh-CN.json'
window.locale('en-US',en_local)
window.locale('zh-CN',zh_local)
