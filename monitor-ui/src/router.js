import Vue from 'vue'
import Router from 'vue-router'

Vue.use(Router)


const router = new Router({
    routes: [
      {
        path: '/',
        name: 'index',
        component: () => import('@/components/index'),
        redirect: '/portal',
        title: '测试首页',
        children: [
          { path: 'alarmManagement',  name: 'alarmManagement', title: '告警管理', meta: {},
           component: () => import('@/components/pages/alarm-management') },
          { path: 'portal',  name: 'portal', title: '首页', meta: {},
           component: () => import('@/components/pages/portal') },
          { path: 'mainView',  name: 'mainView', title: '对象监控', meta: {},
           component: () => import('@/components/pages/main-view') },
          { path: 'monitorConfigIndex',  name: 'monitorConfigIndex', title: '', meta: {},
            component: () => import('@/components/pages/monitor-config/Monitor-config-index'),
            redirect: '/monitorConfigIndex/objectManagement', 
            children: [
              { path: 'objectManagement',  name: 'objectManagement', title: '对象管理', meta: {},
              component: () => import('@/components/pages/monitor-config/Object-management') },
              { path: 'groupManagement',  name: 'groupManagement', title: '组管理', meta: {},
              component: () => import('@/components/pages/monitor-config/Group-management') },
              { path: 'thresholdManagement',  name: 'thresholdManagement', title: '阀值配置', meta: {},
              component: () => import('@/components/pages/monitor-config/Threshold-management') },
            ]
          },
          { path: 'metricConfig',  name: 'metricConfig', title: '视图配置', meta: {},
           component: () => import('@/components/pages/metric-config') },
          { path: 'viewConfigIndex',  name: 'viewConfigIndex', title: '自定义视图主页', meta: {},
           component: () => import('@/components/pages/DIY-view/view-config-index') },
          { path: 'viewConfig',  name: 'viewConfig', title: '自定义视图', meta: {},
           component: () => import('@/components/pages/DIY-view/view-config') },
          { path: 'editView',  name: 'editView', title: '自定义视图编辑', meta: {},
           component: () => import('@/components/pages/DIY-view/edit-view') },
        ]
      },
      {
        path: '/test',
        name: 'test',
        component: () => import('@/components/test'),
        title: 'test'
      }
    ]  
}) 

export default router