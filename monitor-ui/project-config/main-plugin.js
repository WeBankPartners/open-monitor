import Vue from 'vue'
import store from '@/store.js'
import router from './router-plugin'
import routerP from './router-plugin-p'
import 'bootstrap/dist/js/bootstrap.min.js'
import 'font-awesome/css/font-awesome.css'
import '@/plugins/iview.js'
import httpRequestEntrance from '@/assets/js/httpRequestEntrance.js'
import jquery from 'jquery'
import {tableUtil} from '@/assets/js/tableUtil.js'
import {validate} from '@/assets/js/validate.js'
import VeeValidate, {veeValidateConfig} from '@/assets/veeValidate/VeeValidate'
import { Validator } from 'vee-validate'
import apiCenter from '@/assets/config/api-center.json'
import vuex from 'vuex'
import TagShow from '@/components/Tag-show'
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
    childBreadcrumb: { 'zh-CN': '对象', 'en-US': 'Endpoint' }
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
    childBreadcrumb: { 'zh-CN': '关键字', 'en-US': 'Keyword' }
  },
  'monitorConfigIndex/resourceLevel': {
    parentBreadcrumb: {'zh-CN': '对象', 'en-US': 'Object'},
    childBreadcrumb: { 'zh-CN': '层级对象', 'en-US': 'Resource Level' }
  },
  'monitorConfigIndex/businessMonitor': {
    parentBreadcrumb: {'zh-CN': '指标', 'en-US': 'Quota'},
    childBreadcrumb: { 'zh-CN': '业务配置', 'en-US': 'Business Configuration' }
  },
  'monitorConfigIndex/logTemplate': {
    parentBreadcrumb: {'zh-CN': '指标', 'en-US': 'Quota'},
    childBreadcrumb: { 'zh-CN': '业务日志模板', 'en-US': 'Log Template' }
  },
  'monitorConfigIndex/metricConfig': {
    parentBreadcrumb: {'zh-CN': '指标', 'en-US': 'Quota'},
    childBreadcrumb: { 'zh-CN': '指标列表', 'en-US': 'Metric List' }
  },
  'viewConfig': {
    parentBreadcrumb: {'zh-CN': '监控', 'en-US': 'Monitor'},
    childBreadcrumb: { 'zh-CN': '查看视图', 'en-US': 'View' }
  },
  'viewChart': {
    parentBreadcrumb: {'zh-CN': '监控', 'en-US': 'Monitor'},
    childBreadcrumb: { 'zh-CN': '视图', 'en-US': 'View' }
  },
  'alarmHistory': {
    parentBreadcrumb: {'zh-CN': '监控', 'en-US': 'Monitor'},
    childBreadcrumb: { 'zh-CN': '历史告警', 'en-US': 'History Alarm' }
  },
  'viewConfigIndex/boardList': {
    parentBreadcrumb: {'zh-CN': '看板', 'en-US': 'Board'},
    childBreadcrumb: { 'zh-CN': '列表', 'en-US': 'List' }
  },
  'viewConfigIndex/allChartList': {
    parentBreadcrumb: {'zh-CN': '图表库', 'en-US': 'Chart Library'},
    childBreadcrumb: { 'zh-CN': '列表', 'en-US': 'List' }
  },
  'adminConfig/typeConfig': {
    parentBreadcrumb: {'zh-CN': '基础类型', 'en-US': 'Basic Type'},
    childBreadcrumb: { 'zh-CN': '类型配置', 'en-US': 'Type Config' }
  },
  'adminConfig/adminMetric': {
    parentBreadcrumb: {'zh-CN': '基础类型', 'en-US': 'Basic Type'},
    childBreadcrumb: { 'zh-CN': '指标配置', 'en-US': 'Metric Config' }
  },
  'adminConfig/groupBoard': {
    parentBreadcrumb: {'zh-CN': '基础类型', 'en-US': 'Basic Type'},
    childBreadcrumb: { 'zh-CN': '看板配置', 'en-US': 'Board Config' }
  },
  'adminConfig/exporter': {
    parentBreadcrumb: {'zh-CN': '其他', 'en-US': 'Other'},
    childBreadcrumb: { 'zh-CN': '采集器', 'en-US': 'Exporter'}
  },
  'adminConfig/remoteSync': {
    parentBreadcrumb: {'zh-CN': '其他', 'en-US': 'Other'},
    childBreadcrumb: { 'zh-CN': '远程同步', 'en-US': 'Remote Sync'}
  },
  'adminConfig/prometheusLogs': {
    parentBreadcrumb: {'zh-CN': '其他', 'en-US': 'Other'},
    childBreadcrumb: { 'zh-CN': 'Prometheus日志', 'en-US': 'Prometheus Logs'}
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
window.component('TagShow', TagShow)
window.use(VeeValidate, veeValidateConfig)

const ensureRootValidator = () => {
  const vm = window.vm
  if (!vm || (vm.$validator && typeof vm.$validator.validate === 'function')) {
    return
  }
  vm.$validator = new Validator(null, { fastExit: true })
  vm.$nextTick(() => vm.$forceUpdate())
}

if (document.readyState === 'complete') {
  ensureRootValidator()
} else {
  window.addEventListener('load', ensureRootValidator)
}

import DelConfirm from '@/components/del-confirm/index.js'
window.use(DelConfirm)


import Dashboard from '@/views/dashboard'
window.addHomepageComponent && window.addHomepageComponent({
  code: 'MONITOR_CUSTOM_DASHBOARD',
  name: () => {
    return window.vm.$t('m_menu_homepageName')
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
