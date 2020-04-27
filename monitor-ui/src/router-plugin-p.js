
// import alarmManagement from "@/views/alarm-management";
// import dashboard from "@/views/dashboard";
// import endpointView from "@/views/endpoint-view";
// import monitorConfigIndex from "@/views/monitor-config/monitor-config-index";
// import endpointManagement from "@/views/monitor-config/endpoint-management";
// import groupManagement from "@/views/monitor-config/group-management";
// import thresholdManagement from "@/views/monitor-config/threshold-management";
// import logManagement from "@/views/monitor-config/log-management";
// import metricConfig from "@/views/metric-config";
// import viewConfigIndex from "@/views/custom-view/view-config-index";
// import viewConfig from "@/views/custom-view/view-config";
// import editView from "@/views/custom-view/edit-view";
// import portal from "@/views/portal";
// import index from "@/views/index";

const routerP = [
  { path: "/index",
   name: "index", 
  //  title: "首页", 
  //  meta: {}, 
  //  component: index
   },
  {
    path: "/dashboard",
    name: "dashboard",
    // title: "首页",
    // meta: {},
    // component: dashboard
  },
  {
    path: "/viewConfig",
    name: "viewConfig",
    // title: "自定义视图",
    // meta: {},
    // component: viewConfig
  },
  {
    path: "/editPieView",
    name: "editPieView",
    // title: "自定义视图编辑",
    // meta: {},
    // component: editView
  },
  {
    path: "/editLineView",
    name: "editLineView",
    // title: "自定义视图编辑",
    // meta: {},
    // component: editView
  },
  {
    path: "/viewChart",
    name: "viewChart",
    // title: "自定义视图放大",
    // meta: {},
    // component: viewChart
  },
  {
    path: "/portal",
    name: "portal",
    // title: "搜索主页",
    // meta: {},
    // component: portal
  }
];

export default routerP;
