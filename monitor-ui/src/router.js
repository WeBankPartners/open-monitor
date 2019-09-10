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