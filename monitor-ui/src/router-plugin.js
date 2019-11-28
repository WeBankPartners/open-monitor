import Vue from "vue";
import Router from "vue-router";

import alarmManagement from "@/views/alarm-management";
import portal from "@/views/portal";
import endpointView from "@/views/endpoint-view";
import monitorConfigIndex from "@/views/monitor-config/monitor-config-index";
import endpointManagement from "@/views/monitor-config/endpoint-management";
import groupManagement from "@/views/monitor-config/group-management";
import thresholdManagement from "@/views/monitor-config/threshold-management";
import logManagement from "@/views/monitor-config/log-management";
import metricConfig from "@/views/metric-config";
import viewConfigIndex from "@/views/custom-view/view-config-index";
import viewConfig from "@/views/custom-view/view-config";
import editView from "@/views/custom-view/edit-view";
import searchHomepage from "@/views/search-homepage";
import index from "@/view/index";

Vue.use(Router);
const router = [
  { path: "/index", name: "index", title: "首页", meta: {}, component: index },
  {
    path: "/alarmManagement",
    name: "alarmManagement",
    title: "告警管理",
    meta: {},
    component: alarmManagement
  },
  {
    path: "/portal",
    name: "portal",
    title: "首页",
    meta: {},
    component: portal
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
        title: "阀值配置",
        meta: {},
        component: thresholdManagement
      },
      {
        path: "logManagement",
        name: "logManagement",
        title: "阀值配置",
        meta: {},
        component: logManagement
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
    path: "/viewConfigIndex",
    name: "viewConfigIndex",
    title: "自定义视图主页",
    meta: {},
    component: viewConfigIndex
  },
  {
    path: "/viewConfig",
    name: "viewConfig",
    title: "自定义视图",
    meta: {},
    component: viewConfig
  },
  {
    path: "/editView",
    name: "editView",
    title: "自定义视图编辑",
    meta: {},
    component: editView
  },
  {
    path: "/searchHomepage",
    name: "searchHomepage",
    title: "搜索主页",
    meta: {},
    component: searchHomepage
  }
];

export default router;
