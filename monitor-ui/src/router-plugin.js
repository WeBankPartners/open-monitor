import Vue from 'vue'
import Router from 'vue-router'

import alarmManagement from '@/components/pages/alarm-management'
import portal from '@/components/pages/portal'
import mainView from '@/components/pages/main-view'
import monitorConfigIndex from '@/components/pages/monitor-config/Monitor-config-index'
import endpointManagement from '@/components/pages/monitor-config/Endpoint-management'
import groupManagement from '@/components/pages/monitor-config/Group-management'
import thresholdManagement from '@/components/pages/monitor-config/Threshold-management'
import logManagement from '@/components/pages/monitor-config/Log-management'
import metricConfig from '@/components/pages/metric-config'
import viewConfigIndex from '@/components/pages/DIY-view/view-config-index'
import viewConfig from '@/components/pages/DIY-view/view-config'
import editView from '@/components/pages/DIY-view/edit-view'
import searchHomepage from '@/components/pages/Search-homepage'
import index from '@/components/index'

Vue.use(Router)
const router = [
  { path: 'index',  name: 'index', title: '首页', meta: {},
    component: index },
  { path: 'alarmManagement',  name: 'alarmManagement', title: '告警管理', meta: {},
    component: alarmManagement },
  { path: 'portal',  name: 'portal', title: '首页', meta: {},
    component: portal },
  { path: 'mainView',  name: 'mainView', title: '对象监控', meta: {},
    component: mainView },
  { path: 'monitorConfigIndex',  name: 'monitorConfigIndex', title: '', meta: {},
    component: monitorConfigIndex,
    redirect: 'monitorConfigIndex/endpointManagement', 
    children: [
      { path: 'endpointManagement',  name: 'endpointManagement', title: '对象管理', meta: {},
      component: endpointManagement },
      { path: 'groupManagement',  name: 'groupManagement', title: '组管理', meta: {},
      component: groupManagement },
      { path: 'thresholdManagement',  name: 'thresholdManagement', title: '阀值配置', meta: {},
      component: thresholdManagement },
      { path: 'logManagement',  name: 'logManagement', title: '阀值配置', meta: {},
      component: logManagement },
    ]
  },
  { path: 'metricConfig',  name: 'metricConfig', title: '视图配置', meta: {},
    component: metricConfig },
  { path: 'viewConfigIndex',  name: 'viewConfigIndex', title: '自定义视图主页', meta: {},
    component: viewConfigIndex },
  { path: 'viewConfig',  name: 'viewConfig', title: '自定义视图', meta: {},
    component: viewConfig },
  { path: 'editView',  name: 'editView', title: '自定义视图编辑', meta: {},
    component: editView },
    { path: 'searchHomepage',  name: 'searchHomepage', title: '搜索主页', meta: {},
    component: searchHomepage }
]

export default router