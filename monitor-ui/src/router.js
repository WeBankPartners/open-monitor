import Vue from "vue";
import Router from "vue-router";
import {cookies} from '@/assets/js/cookieUtils';

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
            import("@/views/monitor-config/monitor-config-index"),
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
              title: "组管理",
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
          path: "metricConfig",
          name: "metricConfig",
          title: "视图配置",
          meta: {},
          component: () => import("@/views/metric-config")
        },
        {
          path: "viewConfigIndex",
          name: "viewConfigIndex",
          title: "自定义视图主页",
          meta: {},
          component: () =>
            import("@/views/custom-view/view-config-index")
        },
        {
          path: "viewConfig",
          name: "viewConfig",
          title: "自定义视图",
          meta: {},
          component: () => import("@/views/custom-view/view-config")
        },
        {
          path: "editView",
          name: "editView",
          title: "自定义视图编辑",
          meta: {},
          component: () => import("@/views/custom-view/edit-view")
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
      path: "/test",
      name: "test",
      component: () => import("@/views/test"),
      title: "test"
    }
  ]
});

router.beforeEach((to, from, next) => {
  if (!cookies.getCookie('Authorization')&& to.name != 'login'&& to.name != 'register') {
    next({name:'login'})
  } else {
    next()
  }
})
export default router;
