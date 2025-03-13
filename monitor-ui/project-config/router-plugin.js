
import alarmManagement from "@/views/alarm-management";
// import alarmHistory from "@/views/alarm-history";
import dashboard from "@/views/dashboard";
import endpointView from "@/views/endpoint-view";
import monitorConfigIndex from "./pages/monitor-config-index";
import endpointManagement from "@/views/monitor-config/endpoint-management";
import groupManagement from "@/views/monitor-config/group-management";
import thresholdManagement from "@/views/monitor-config/threshold-management";
import logManagement from "@/views/monitor-config/log-management";
import resourceLevel from "@/views/monitor-config/resource-level";
import businessMonitor from "@/views/monitor-config/business-monitor";
import logTemplate from "@/views/monitor-config/log-template";
import metricConfig from "@/views/metric-config/index";
import viewConfigIndex from "@/views/custom-view/view-config-index";
import chartList from "@/views/custom-view/chart-list";
import boardIndex from "@/views/custom-view/index";
import viewConfig from "@/views/custom-view/view-config";
import viewChart from "@/views/custom-view/view-chart";
import portal from "@/views/portal";
import adminConfigIndex from '@/views/admin-config/index';
import typeConfig from "@/views/admin-config/basic/type-config";
import groupBoard from "@/views/admin-config/basic/board-config";
import adminMetric from '@/views/admin-config/basic/metric-config';
import exporter from "@/views/admin-config/other/exporter";
import remoteSync from "@/views/admin-config/other/remote-sync";
import index from "@/views/index";

const router = [
  { path: "/index", name: "index", title: "首页", meta: {}, component: index },
  {
    path: "/alarmManagement",
    name: "alarmManagement",
    title: "告警管理",
    meta: {},
    component: alarmManagement
  },
  // {
  //   path: "/alarmHistory",
  //   name: "alarmHistory",
  //   title: "告警历史",
  //   meta: {},
  //   component: alarmHistory
  // },
  {
    path: "/dashboard",
    name: "dashboard",
    title: "首页",
    meta: {},
    component: dashboard
  },
  {
    path: "/endpointView",
    name: "endpointView",
    title: "对象监控",
    meta: {},
    component: endpointView
  },
  {
    path: "/monitorConfigIndex",
    name: "monitorConfigIndex",
    title: "",
    meta: {},
    component: monitorConfigIndex,
    redirect: "/monitorConfigIndex/endpointManagement",
    children: [
      {
        path: "endpointManagement",
        name: "endpointManagement",
        title: "对象管理",
        meta: {},
        component: endpointManagement
      },
      {
        path: "groupManagement",
        name: "groupManagement",
        title: "组管理",
        meta: {},
        component: groupManagement
      },
      {
        path: "thresholdManagement",
        name: "thresholdManagement",
        title: "阈值配置",
        meta: {},
        component: thresholdManagement
      },
      {
        path: "logManagement",
        name: "logManagement",
        title: "关键字配置",
        meta: {},
        component: logManagement
      },
      {
        path: "resourceLevel",
        name: "resourceLevel",
        title: "资源层级",
        meta: {},
        component: resourceLevel
      },
      {
        path: "businessMonitor",
        name: "businessMonitor",
        title: "businessMonitor",
        meta: {},
        component: businessMonitor
      },
      {
        path: "logTemplate",
        name: "logTemplate",
        title: "日志模版",
        meta: {},
        component: logTemplate
      },
      {
        path: "metricConfig",
        name: "metricConfig",
        title: "metricConfig",
        meta: {},
        component: metricConfig
      }
    ]
  },
  {
    path: "/metricConfig",
    name: "metricConfig",
    title: "视图配置",
    meta: {},
    component: metricConfig
  },
  {
    path: "viewConfigIndex",
    name: "viewConfigIndex",
    title: "自定义视图主页",
    meta: {},
    component: boardIndex,
    redirect: "/viewConfigIndex/boardList",
    children: [
      {
        path: "boardList",
        name: "boardList",
        title: "看板列表",
        meta: {},
        component: viewConfigIndex
      },
      {
        path: "allChartList",
        name: "allChartList",
        title: "列表",
        meta: {},
        component: chartList
      },
      {
        path: "/viewConfig",
        name: "viewConfig",
        title: "自定义视图",
        meta: {},
        component: viewConfig
      }
    ]
  },
  {
    path: "/viewChart",
    name: "viewChart",
    title: "自定义视图放大",
    meta: {},
    component: viewChart
  },
  {
    path: "/portal",
    name: "portal",
    title: "搜索主页",
    meta: {},
    component: portal
  },
  {
    path: "/adminConfig",
    name: "adminConfig",
    title: "管理员配置",
    meta: {},
    redirect: '/adminConfig/groupBoard',
    component: adminConfigIndex,
    children: [
      {
        path: "typeConfig",
        name: "typeConfig",
        title: "类型配置",
        meta: {},
        component: typeConfig
      },
      {
        path: "groupBoard",
        name: "groupBoard",
        title: "看板配置",
        meta: {},
        component: groupBoard
      },
      {
        path: "adminMetric",
        name: "adminMetric",
        title: "指标配置",
        meta: {},
        component: adminMetric
      },
      {
        path: "exporter",
        name: "exporter",
        title: "exporter",
        meta: {},
        component: exporter
      },
      {
        path: "remoteSync",
        name: "remoteSync",
        title: "remoteSync",
        meta: {},
        component: remoteSync
      }
    ]
  }
];

export default router;
