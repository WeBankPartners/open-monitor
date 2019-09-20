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
          { path: 'portal',  name: 'portal', title: '首页', meta: {},
           component: () => import('@/components/pages/portal') },
          { path: 'mainView',  name: 'mainView', title: '监控视图', meta: {},
           component: () => import('@/components/pages/main-view') },
          { path: 'monitorConfigIndex',  name: 'monitorConfigIndex', title: '监控配置', meta: {},
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