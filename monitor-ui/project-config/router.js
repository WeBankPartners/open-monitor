import Vue from "vue";
import Router from "vue-router";
// import { getToken } from '@/assets/js/cookies.ts'

Vue.use(Router);

const router = new Router({
  scrollBehavior: () => ({ // 滚动条滚动的行为，不加这个默认就会记忆原来滚动条的位置
    y: 0
  }),
  routes: [
    {
      path: "/",
      name: "index",
      component: () => import("@/views/index"),
      redirect: "/dashboard",
      title: "测试首页",
      children: [
        {
          path: "alarmManagement",
          name: "alarmManagement",
          title: "告警管理",
          meta: {},
          component: () => import("@/views/alarm-management")
        },
        {
          path: "alarmHistory",
          name: "alarmHistory",
          title: "告警历史",
          meta: {},
          component: () => import("@/views/alarm-history")
        },
        {
          path: "dashboard",
          name: "dashboard",
          title: "首页",
          meta: {},
          component: () => import("@/views/dashboard")
        },
        {
          path: "systemMonitoring",
          name: "systemMonitoring",
          title: "对象监控",
          meta: {},
          component: () => import("@/views/system-monitoring/system-monitoring")
        },
        {
          path: "sysViewChart",
          name: "sysViewChart",
          title: "视图查看",
          meta: {},
          component: () => import("@/views/system-monitoring/sys-view-chart")
        },
        {
          path: "endpointView",
          name: "endpointView",
          title: "对象监控",
          meta: {},
          component: () => import("@/views/endpoint-view")
        },
        {
          path: "monitorConfigIndex",
          name: "monitorConfigIndex",
          title: "",
          meta: {},
          component: () =>
            import("./pages/monitor-config-index"),
          redirect: "/monitorConfigIndex/endpointManagement",
          children: [
            {
              path: "endpointManagement",
              name: "endpointManagement",
              title: "对象管理",
              meta: {},
              component: () =>
                import("@/views/monitor-config/endpoint-management")
            },
            {
              path: "groupManagement",
              name: "groupManagement",
              title: "对象组",
              meta: {},
              component: () =>
                import("@/views/monitor-config/group-management")
            },
            {
              path: "thresholdManagement",
              name: "thresholdManagement",
              title: "阈值配置",
              meta: {},
              component: () =>
                import("@/views/monitor-config/threshold-management")
            },
            {
              path: "logManagement",
              name: "logManagement",
              title: "关键字配置",
              meta: {},
              component: () =>
                import("@/views/monitor-config/log-management")
            },
            {
              path: "resourceLevel",
              name: "resourceLevel",
              title: "资源层级",
              meta: {},
              component: () =>
                import("@/views/monitor-config/resource-level")
            },
            {
              path: "businessMonitor",
              name: "businessMonitor",
              title: "businessMonitor",
              meta: {},
              component: () =>
                import("@/views/monitor-config/business-monitor")
            },
            {
              path: "logTemplate",
              name: "logTemplate",
              title: "日志模版",
              meta: {},
              component: () => import("@/views/monitor-config/log-template")
            },
            {
              path: "metricConfig",
              name: "metricConfig",
              title: "metricConfig",
              meta: {},
              component: () => import("@/views/metric-config/index")
            }
          ]
        },
        {
          path: "userConfigIndex",
          name: "userConfigIndex",
          title: "",
          meta: {},
          component: () =>
            import("@/views/user-management/user-config-index"),
          redirect: "/userConfigIndex/userInformationModify",
          children: [
            {
              path: "userInformationModify",
              name: "userInformationModify",
              title: "用户信息修改",
              meta: {},
              component: () =>
                import("@/views/user-management/user-information-modify")
            },
            {
              path: "userManagement",
              name: "userManagement",
              title: "用户管理",
              meta: {},
              component: () =>
                import("@/views/user-management/user-management")
            },
            {
              path: "roleManagement",
              name: "roleManagement",
              title: "角色管理",
              meta: {},
              component: () =>
                import("@/views/user-management/role-management")
            }
          ]
        },
        {
          path: "viewConfigIndex",
          name: "viewConfigIndex",
          title: "自定义视图主页",
          meta: {},
          redirect: "/viewConfigIndex/boardList",
          component: () => import("@/views/custom-view/index"),
          children: [
            {
              path: "boardList",
              name: "boardList",
              title: "看板列表",
              meta: {},
              component: () =>
                import("@/views/custom-view/view-config-index")
            },
            {
              path: "allChartList",
              name: "allChartList",
              title: "列表",
              meta: {},
              component: () =>
                import("@/views/custom-view/chart-list")
            },
            {
              path: "viewConfig",
              name: "viewConfig",
              title: "自定义视图",
              meta: {},
              component: () => import("@/views/custom-view/view-config")
            }
          ]
        },
        {
          path: "viewChart",
          name: "viewChart",
          title: "视图查看",
          meta: {},
          component: () => import("@/views/custom-view/view-chart")
        },
        {
          path: "portal",
          name: "portal",
          title: "搜索主页",
          meta: {},
          component: () => import("@/views/portal")
        },
        {
          path: "adminConfig",
          name: "adminConfig",
          title: "管理员配置",
          meta: {},
          redirect: '/adminConfig/groupBoard',
          component: () => import("@/views/admin-config/index"),
          children: [
            {
              path: "typeConfig",
              name: "typeConfig",
              title: "类型配置",
              meta: {},
              component: () => import("@/views/admin-config/basic/type-config")
            },
            {
              path: "groupBoard",
              name: "groupBoard",
              title: "看板配置",
              meta: {},
              component: () => import("@/views/admin-config/basic/board-config")
            },
            {
              path: "adminMetric",
              name: "adminMetric",
              title: "指标配置",
              meta: {},
              component: () => import("@/views/admin-config/basic/metric-config")
            },
            {
              path: "exporter",
              name: "exporter",
              title: "exporter",
              meta: {},
              component: () =>
                import("@/views/admin-config/other/exporter")
            },
            {
              path: "remoteSync",
              name: "remoteSync",
              title: "remoteSync",
              meta: {},
              component: () =>
                import("@/views/admin-config/other/remote-sync")
            }
          ]
        }
      ]
    },
    {
      path: "/login",
      name: "login",
      component: () => import("@/views/login"),
      title: "login"
    },
    {
      path: "/register",
      name: "register",
      component: () => import("@/views/register"),
      title: "register"
    },
    {
      path: "/endpointViewExternalCall",
      name: "endpointViewExternalCall",
      title: "对象监控外链调用",
      meta: {},
      component: () => import("@/views/endpoint-view-external-call")
    },
    {
      path: "/callCustomViewExternal",
      name: "callCustomViewExternal",
      title: "自定义视图外链调用",
      meta: {},
      component: () => import("@/views/call-custom-view-external")
    }
  ]
});

const originalPush = Router.prototype.push
Router.prototype.push = function push(location) {
 return originalPush.call(this, location).catch(err => err)
}

router.beforeEach((to, from, next) => {
  // if (['login', 'callCustomViewExternal'].includes(to.name)) {
  // next()
  //   return
  // }
  // if (!getToken()&& !['login', 'register', 'endpointViewExternalCall', 'callCustomViewExternal'].includes(to.name)) {
  //   next({name:'login'})
  // } else {
  //   next()
  // }
  next()
})
export default router;
